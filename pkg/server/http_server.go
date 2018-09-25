package server

import (
	"log"
	"net/http"

	"github.com/adjust/rmq"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/s2iservice/pkg/httphandler"
	"github.com/s2iservice/pkg/logger"
)

type HttpServer struct {
	Db         *mongo.Database
	Connection rmq.Connection
	Queue      rmq.Queue
	Jobs       *httphandler.JobService
	// Logs       *httphandler.LogService
	// Templates  *httphandler.TemplateService
}

const APIVersion = "/api/v1alpha1"

func (s *HttpServer) GetRouter() rest.App {
	router, err := rest.MakeRouter(
		rest.Get("/jobs", s.Jobs.GetJobsHandler),
		rest.Post("/jobs", s.Jobs.AddJobHandler),
		rest.Get("/jobs/:jid", s.Jobs.GetJobHandler),
		rest.Post("/jobs/:jid", s.Jobs.UpdateJobHandler),
		rest.Post("/jobs/:jid/run", s.Jobs.RunJobHandler),
		// rest.Get("/jobs/:jid/logs", s.Logs.GetLoggerHandler),
		// rest.Post("/templates", s.Templates.CreateTemplatesHandler),
		// rest.Get("/templates", s.Templates.GetAvailableTemplatesHandler),
	)
	if err != nil {
		logger.Critical("%v", err)
		return nil
	}
	return router
}

func (s *HttpServer) Serve(r *Resources) {
	s.Db = r.Db
	s.Queue = r.Redis
	s.Connection = r.Connection
	s.Jobs = httphandler.NewJobService(r.Db, r.Redis)
	// s.Logs = httphandler.NewLogService(r.Db)
	// s.Templates = httphandler.NewTemplateService(r.Db)
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			return true
		},
	})
	api.SetApp(s.GetRouter())
	http.Handle(APIVersion+"/", http.StripPrefix(APIVersion, api.MakeHandler()))
	http.Handle("/overview", NewHandler(s.Connection))
	log.Fatal(http.ListenAndServe(":8001", nil))
}
