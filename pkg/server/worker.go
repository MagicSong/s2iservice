package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/adjust/rmq"
	"github.com/go-redis/redis"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
	"github.com/s2iservice/pkg/constants"
	"github.com/s2iservice/pkg/docker"
	"github.com/s2iservice/pkg/logger"
	"github.com/s2iservice/pkg/models"
)

const (
	MaxCount      = 4
	CMD           = "s2i"
	MaxRetry      = 2
	CleanInterval = 2
)

type Worker struct {
	Db                        *mongo.Database
	Queue                     rmq.Queue
	MaxConsumerCountPerSecond int
}

func (w *Worker) Work(r *Resources) {
	w.Db = r.Db
	w.Queue = r.Redis
	if w.MaxConsumerCountPerSecond == 0 {
		w.MaxConsumerCountPerSecond = MaxCount
	}
	w.Queue.StartConsuming(w.MaxConsumerCountPerSecond, time.Second)
	for index := 0; index < MaxCount; index++ {
		name := fmt.Sprintf("consumer %d", index)
		w.Queue.AddConsumer(name, w.NewConsumer(index))
	}
	logger.Info("s2i start to watch tasks queue")
	go func() {
		client := redis.NewClient(&redis.Options{
			Addr:     r.cfg.Redis.Address,
			Password: r.cfg.Redis.Password, // no password set
			DB:       r.cfg.Redis.DB,       // use default DB
		})
		connection := rmq.OpenConnectionWithRedisClient(r.cfg.Redis.RMQName, client)
		cleaner := rmq.NewCleaner(connection)
		for range time.Tick(CleanInterval * time.Minute) {
			logger.Info("Begin to clean Queue")
			cleaner.Clean()
			w.Queue.ReturnAllRejected()
		}
	}()
	select {}
}

type S2IConsumer struct {
	name   string
	count  int
	before time.Time
	job    *models.S2IJob
	worker *Worker
}

var (
	mutex sync.Mutex
)

func (w *Worker) NewConsumer(tag int) *S2IConsumer {
	return &S2IConsumer{
		name:   fmt.Sprintf("consumer%d", tag),
		count:  0,
		before: time.Now(),
		worker: w,
	}
}

func (s *S2IConsumer) getJob(jid string, username string) (*models.S2IJob, error) {
	job := &models.S2IJob{}
	filter := bson.NewDocument(bson.EC.String("_id", jid), bson.EC.String("username", username))
	err := s.worker.Db.Collection(constants.S2IJobCollectionName).FindOne(context.Background(), filter).Decode(job)
	if err != nil {
		return nil, err
	}
	return job, nil
}
func (s *S2IConsumer) Consume(delivery rmq.Delivery) {
	s.before = time.Now()
	coll := s.worker.Db.Collection(constants.S2IJobCollectionName)
	var redisJob models.RedisJob
	if err := json.Unmarshal([]byte(delivery.Payload()), &redisJob); err != nil {
		glog.Errorf("%s", err.Error())
		delivery.Reject()
		return
	}

	job, err := s.getJob(redisJob.ID, redisJob.Username)
	if err != nil {
		logger.Error("Detected noexit or deleted job <%v>!\nError: %v", redisJob, err)
		delivery.Ack()
		return
	}
	s.job = job
	// perform task
	cmd := exec.Command(CMD, job.Parameters...)
	stdout, _ := cmd.StderrPipe()
	done := make(chan struct{})
	errChan := make(chan error)
	defer close(errChan)
	defer close(done)
	go s.do(cmd, stdout, done, errChan)
	logger.Info("cusumer <%s> performing task <%s>, current consumer has processed %d tasks", s.name, job.ID, s.count)
	updateJobStatusAndInfo(coll, s.job, models.Processing, "Under Processing")
	select {
	case <-done:
		delivery.Ack()
	case e := <-errChan:
		logger.Error("There was error in task %s, Error: %v", job.ID, e)
		updateJobStatusAndInfo(coll, s.job, models.Error, e.Error())
		err = incremRetry(coll, s.job.ID)
		if err != nil {
			glog.Errorf("%s", err.Error())
		}
		delivery.Reject()
	}
	s.count++
}

func (s *S2IConsumer) do(cmd *exec.Cmd, out io.ReadCloser, done chan<- struct{}, errChan chan error) {
	coll := s.worker.Db.Collection(constants.S2IJobCollectionName)
	retry, err := getRetry(coll, s.job.ID)
	if err != nil {
		errChan <- err
		return
	}
	if retry >= MaxRetry {
		done <- struct{}{}
		logger.Error("task %s is TERMINATING because of exceeding the retry limit", s.job.ID)
		updateJobStatusAndInfo(coll, s.job, models.Terminated, "Terminated due to exceeding retry limit")
		return
	}
	s.job.Retry = retry
	err = cmd.Start()
	if err != nil {
		errChan <- err
		return
	}
	err = s.storeOutputIntoDatabase(out)
	if err != nil {
		errChan <- err
	}
	err = cmd.Wait()
	if err != nil {
		errChan <- err
	} else {
		if !s.job.Export {
			done <- struct{}{}
			logger.Info("task %s is done", s.job)
			updateJobStatusAndInfo(coll, s.job, models.Completed, "Done")
		} else {
			logger.Info("(job-id:%s) try to push image to remote registry", s.job.ID)
			err = docker.PushImage(s.job.PushUsername, s.job.PushPassword, s.job.ImageName, s.storeOutputIntoDatabase)
			if err != nil {
				errChan <- err
				return
			}
			done <- struct{}{}
			logger.Info("task %s is done", s.job)
			updateJobStatusAndInfo(coll, s.job, models.Completed, "Done")
		}
	}
}
func (s *S2IConsumer) storeOutputIntoDatabase(out io.ReadCloser) (err error) {
	id := 0
	start := time.Now()
	rows := make([]interface{}, 0)
	scanner := bufio.NewScanner(out)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		logger.Info("(job-id:%s) %s", s.job.ID, m)
		if strings.Contains(m, "errorDetail") {
			return fmt.Errorf("%s", m)
		}
		id++
		rows = append(rows, &models.LogRow{
			JobID:   s.job.ID,
			Text:    m,
			LogTime: time.Now(),
			RetryID: s.job.Retry,
			Seq:     id,
		})
		if time.Now().Sub(start) >= 5*time.Second {
			_, err = s.worker.Db.Collection(constants.S2ILogCollectionName).InsertMany(context.Background(), rows)
			if err != nil {
				return
			}
			rows = nil
			start = time.Now()
		}
	}
	if len(rows) > 0 {
		_, err = s.worker.Db.Collection(constants.S2ILogCollectionName).InsertMany(context.Background(), rows)
		return err
	}
	return
}
func updateJobStatusAndInfo(db *mongo.Collection, job *models.S2IJob, status models.JobStatus, info string) error {
	filter := bson.NewDocument(bson.EC.String("_id", job.ID))
	update := bson.NewDocument(bson.EC.SubDocumentFromElements("$set", bson.EC.String("status", string(status)), bson.EC.String("info", info), bson.EC.Time("update_time", time.Now())))
	_, err := db.UpdateOne(context.TODO(), filter, update)
	return err
}

func getRetry(c *mongo.Collection, jobID string) (uint8, error) {
	var result struct {
		Retry uint8
	}
	filter := bson.NewDocument(bson.EC.String("_id", jobID))
	projection := findopt.Projection(bson.NewDocument(bson.EC.Int32("retry", 1)))
	err := c.FindOne(context.Background(), filter, projection).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Retry, nil
}
func incremRetry(c *mongo.Collection, jobID string) (err error) {
	filter := bson.NewDocument(bson.EC.String("_id", jobID))
	update := bson.NewDocument(bson.EC.SubDocumentFromElements("$inc", bson.EC.Int32("retry", 1)))
	_, err = c.UpdateOne(context.Background(), filter, update)
	return
}
