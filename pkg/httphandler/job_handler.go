package httphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/adjust/rmq"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/bsoncodec"
	"github.com/mongodb/mongo-go-driver/mongo"

	"github.com/s2iservice/pkg/api"
	"github.com/s2iservice/pkg/constants"
	"github.com/s2iservice/pkg/utils/idutils"
)

type JobService struct {
	Db    *mongo.Database
	Queue rmq.Queue
}

func NewJobService(db *mongo.Database, q rmq.Queue) *JobService {
	return &JobService{
		Db:    db,
		Queue: q,
	}
}

func (s *JobService) getJob(jid string, username string) (*api.S2IJob, error) {
	job := &api.S2IJob{}
	filter := bson.NewDocument(bson.EC.String("_id", jid), bson.EC.String("username", username))
	err := s.Db.Collection(constants.S2IJobCollectionName).FindOne(context.Background(), filter).Decode(job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *JobService) GetJobHandler(w rest.ResponseWriter, r *rest.Request) {
	jid, _ := r.PathParams["jid"]
	username := r.Env["REMOTE_USER"].(string)
	job, err := s.getJob(jid, username)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(job)
}

func (s *JobService) GetJobsHandler(w rest.ResponseWriter, r *rest.Request) {
	username := r.Env["REMOTE_USER"].(string)
	filter := bson.NewDocument(bson.EC.String("username", username))
	cur, err := s.Db.Collection(constants.S2IJobCollectionName).Find(context.Background(), filter)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())
	jobs := make([]*api.S2IJob, 0)
	for cur.Next(context.Background()) {
		job := new(api.S2IJob)
		err := cur.Decode(job)
		if err != nil {
			glog.Errorf("%s", err.Error())
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, job)
	}
	if err := cur.Err(); err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&jobs)
}
func (s *JobService) AddJobHandler(w rest.ResponseWriter, r *rest.Request) {
	job := &api.S2IJob{}
	req := &api.Config{}
	job.Username = r.Env["REMOTE_USER"].(string)
	err := r.DecodeJsonPayload(req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job.Config = req
	job.CreateTime = time.Now()
	job.UpdateTime = time.Now()
	job.ID = idutils.GetUuid(constants.S2IJobIDPrefix)
	_, err = s.Db.Collection(constants.S2IJobCollectionName).InsertOne(context.Background(), job)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.PublishJob(job.ID, job.Username)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(job)
}

func (s *JobService) UpdateJobHandler(w rest.ResponseWriter, r *rest.Request) {
	req := &api.Config{}
	username := r.Env["REMOTE_USER"].(string)
	jid := r.PathParams["jid"]
	job, err := s.getJob(jid, username)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = r.DecodeJsonPayload(req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job.Config = req
	filter := bson.NewDocument(bson.EC.String("_id", jid), bson.EC.String("username", username))
	update := bson.NewDocument(bson.EC.SubDocumentFromElements("$set", bsoncodec.ConstructElement("Config", job), bson.EC.Time("update_time", time.Now())))
	result := s.Db.Collection(constants.S2IJobCollectionName).FindOneAndUpdate(context.Background(), filter, update)
	if result == nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	doc := bson.NewDocument()
	err = result.Decode(doc)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(job)
}

func (s *JobService) PublishJob(jobid, username string) error {
	runID := idutils.GetUuid(constants.S2IRunIDPrefix)
	run := &api.S2IRun{
		JobID:     jobid,
		RunID:     runID,
		StartTime: time.Now(),
		Status:    api.Created,
	}
	_, err := s.Db.Collection(constants.S2IRunCollectionName).InsertOne(context.Background(), run)
	if err != nil {
		return err
	}
	job := &api.RedisJob{
		Username: username,
		RunID:    runID,
		JobID:    jobid,
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		return err
	}
	if ok := s.Queue.PublishBytes(jobBytes); !ok {
		return fmt.Errorf("Failed to publish job")
	}
	return nil
}
func (s *JobService) RunJobHandler(w rest.ResponseWriter, r *rest.Request) {
	jobid := r.PathParams["jid"]
	username := r.Env["REMOTE_USER"].(string)
	_, err := s.getJob(jobid, username)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = s.PublishJob(jobid, username)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *JobService) GetRunHandler(w rest.ResponseWriter, r *rest.Request) {
	jobid := r.PathParams["jid"]
	runid := r.PathParams["runid"]
	username := r.Env["REMOTE_USER"].(string)
	filter := bson.NewDocument(bson.EC.String("_id", runid), bson.EC.String("username", username), bson.EC.String("job_id", jobid))
	result := &api.S2IRun{}
	err := s.Db.Collection(constants.S2IRunCollectionName).FindOne(context.Background(), filter).Decode(result)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(result)
}
func (s *JobService) GetRunsHandler(w rest.ResponseWriter, r *rest.Request) {
	jobid := r.PathParams["jid"]
	username := r.Env["REMOTE_USER"].(string)
	filter := bson.NewDocument(bson.EC.String("username", username), bson.EC.String("job_id", jobid))
	cur, err := s.Db.Collection(constants.S2IRunCollectionName).Find(context.Background(), filter)
	if err != nil {
		glog.Errorf("%s", err.Error())
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())
	runs := make([]*api.S2IRun, 0)
	for cur.Next(context.Background()) {
		run := new(api.S2IRun)
		err := cur.Decode(run)
		if err != nil {
			glog.Errorf("%s", err.Error())
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		runs = append(runs, run)
	}
	if err := cur.Err(); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&runs)
}
