package httphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/s2iservice/pkg/utils/idutils"

	"github.com/adjust/rmq"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/docker/distribution/reference"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/s2iservice/pkg/constants"
	"github.com/s2iservice/pkg/logger"
	"github.com/s2iservice/pkg/models"
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
func GenerateS2IParameters(req *models.S2IRequest) ([]string, error) {
	parameters := make([]string, 4)
	if req.Custom != "" {
		parameters[0] = req.Custom
		parameters[1] = "2>&1"
		return parameters, nil
	}
	parameters[0] = "build"
	_, err := url.Parse(req.SourceURL)
	if err != nil {
		return nil, ErrInvalidGitURL
	}
	parameters[1] = req.SourceURL
	if _, err := reference.ParseNamed(req.BuilderImage); err != nil {
		return nil, ErrInvaildImageName
	}
	parameters[2] = req.BuilderImage
	if _, err := reference.ParseNamed(req.Tag); err != nil {
		return nil, ErrInvaildImageName
	}
	parameters[3] = req.Tag
	if req.CallbackURL != "" {
		_, err := url.Parse(req.CallbackURL)
		if err != nil {
			return nil, ErrInvalidCallbackURL
		}
		parameters = append(parameters, "--callback-url "+req.CallbackURL)
	}
	if req.ContextDir != "" {
		parameters = append(parameters, "--context-dir "+req.ContextDir)
	}
	if req.AddHost != "" {
		parameters = append(parameters, "--add-host "+req.AddHost)
	}
	if req.RuntimeImage != "" {
		if _, err := reference.ParseNamed(req.RuntimeImage); err != nil {
			return nil, ErrInvaildImageName
		}
		parameters = append(parameters, "--runtime-image "+req.RuntimeImage)
		if req.RuntimeArtifact == "" {
			return nil, ErrLackOfRuntimeOption
		}
		parameters = append(parameters, "--runtime-artifact "+req.RuntimeArtifact)
	}
	if req.EnvironmentVariables != "" {
		parameters = append(parameters, "-e "+req.EnvironmentVariables)
	}
	if req.ReuseMavenLocalRepo {
		parameters = append(parameters, "-v")
		parameters = append(parameters, "/tmp/.m2:/opt/app-root/src/.m2")
	}
	parameters = append(parameters, "2>&1") //所有信息都重定向到stdout
	return parameters, nil
}

func (s *JobService) UpdateJobHandler(w rest.ResponseWriter, r *rest.Request) {
	req := &models.S2IRequest{}
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
	parameters, err := GenerateS2IParameters(req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job.Parameters = parameters
	filter := bson.NewDocument(bson.EC.String("_id", jid), bson.EC.String("username", username))
	values := make([]*bson.Value, 0)
	for _, item := range parameters {
		values = append(values, bson.VC.String(item))
	}
	update := bson.NewDocument(bson.EC.SubDocumentFromElements("$set", bson.EC.ArrayFromElements("parameters", values...), bson.EC.Time("update_time", time.Now())))
	result := s.Db.Collection(constants.S2IJobCollectionName).FindOneAndUpdate(context.Background(), filter, update)
	if result == nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	doc := bson.NewDocument()
	err = result.Decode(doc)
	if err != nil {
		logger.Error("error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(job)
}
func (s *JobService) getJob(jid string, username string) (*models.S2IJob, error) {
	job := &models.S2IJob{}
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
	jobs := make([]*models.S2IJob, 0)
	for cur.Next(context.Background()) {
		job := new(models.S2IJob)
		err := cur.Decode(job)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, job)
	}
	if err := cur.Err(); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&jobs)
}
func (s *JobService) AddJobHandler(w rest.ResponseWriter, r *rest.Request) {
	req := &models.S2IRequest{}
	job := &models.S2IJob{}
	job.Username = r.Env["REMOTE_USER"].(string)
	err := r.DecodeJsonPayload(req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	parameters, err := GenerateS2IParameters(req)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job.Parameters = parameters
	job.ImageName = req.Tag
	job.CreateTime = time.Now()
	job.UpdateTime = time.Now()
	if req.Export {
		job.Export = true
		job.PushUsername = req.PushUsername
		job.PushPassword = req.PushPassword
	}
	job.ID = idutils.GetUuid(constants.S2IJobIDPrefix)
	_, err = s.Db.Collection(constants.S2IJobCollectionName).InsertOne(context.Background(), job)
	if err != nil {
		logger.Error("error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redisJob := &models.RedisJob{
		ID:       job.ID,
		Username: job.Username,
	}
	err = PublishJob(redisJob, s.Queue)
	if err != nil {
		logger.Error("error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(job)
}

func PublishJob(job *models.RedisJob, q rmq.Queue) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return err
	}
	if ok := q.PublishBytes(jobBytes); !ok {
		return fmt.Errorf("Failed to publish job")
	}
	return nil
}
func (s *JobService) RunJobHandler(w rest.ResponseWriter, r *rest.Request) {
	job := &models.RedisJob{
		ID:       r.PathParams["jid"],
		Username: r.Env["REMOTE_USER"].(string),
	}
	err := PublishJob(job, s.Queue)
	if err != nil {
		logger.Error("error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
