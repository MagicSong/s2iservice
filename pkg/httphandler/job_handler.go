package httphandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/s2iservice/pkg/utils/idutils"

	"github.com/adjust/rmq"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/docker/distribution/reference"
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

// func (s *JobService) UpdateJobHandler(w rest.ResponseWriter, r *rest.Request) {
// 	req := &models.S2IRequest{}
// 	username := r.Env["REMOTE_USER"].(string)
// 	jid, err := strconv.Atoi(r.PathParams["jid"])
// 	if err != nil {
// 		rest.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	job, err := s.getJob(jid, username)
// 	if err != nil {
// 		rest.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}
// 	err = r.DecodeJsonPayload(req)
// 	if err != nil {
// 		rest.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	parameters, err := GenerateS2IParameters(req)
// 	if err != nil {
// 		rest.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	job.Parameters = parameters
// 	sql := `UPDATE s2ijob
// 	SET parameters=?
// 	WHERE id=? `
// 	_, err = s.Db.UpdateBySql(sql, strings.Join(parameters, " "), jid, username).Exec()
// 	if err != nil {
// 		logger.Error("error: %v", err)
// 		rest.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteJson(job)
// }
func (s *JobService) getJob(jid int, username string) (*models.S2IJob, error) {
	job := &models.S2IJob{}
	return job, nil
}
func (s *JobService) GetJobHandler(w rest.ResponseWriter, r *rest.Request) {
	jid, _ := strconv.Atoi(r.PathParams["jid"])
	username := r.Env["REMOTE_USER"].(string)
	job, err := s.getJob(jid, username)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(job)
}

// func (s *JobService) GetJobsHandler(w rest.ResponseWriter, r *rest.Request) {
// 	jobs := make([]models.S2IJob, 0)
// 	username := r.Env["REMOTE_USER"]
// 	sql := `SELECT id,
//     username,
//     status,
//     create_time,
//     update_time,
//     info,
//     retry
// 	FROM s2ijob
// 	WHERE username=?`
// 	_, err := s.Db.SelectBySql(sql, username).Load(&jobs)
// 	if err != nil {
// 		logger.Error("error: %v", err)
// 		rest.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteJson(&jobs)
// }
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
	jobBytes, err := json.Marshal(job)
	if err != nil {
		logger.Error("error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ok := s.Queue.PublishBytes(jobBytes); !ok {
		logger.Error("error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(job)
}

// func (s *JobService) RunJobHandler(w rest.ResponseWriter, r *rest.Request) {
// 	jid, err := strconv.Atoi(r.PathParams["jid"])
// 	if err != nil {
// 		rest.Error(w, "The format of job id is illegal", http.StatusBadRequest)
// 		return
// 	}
// 	username := r.Env["REMOTE_USER"]
// 	sql := `SELECT *
// 	FROM s2ijob
// 	WHERE id=?
// 	AND
// 	username=?`
// 	job := &models.S2IJob{}
// 	err = s.Db.SelectBySql(sql, jid, username).LoadOne(job)
// 	if err != nil {
// 		rest.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}

// }
