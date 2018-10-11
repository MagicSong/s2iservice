package httphandler

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/google/go-github/github"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type TemplateService struct {
	Github *github.Client
	Db     *mongo.Database
}

func NewTemplateService(db *mongo.Database) *TemplateService {
	return &TemplateService{
		Db: db,
	}
}

func (t *TemplateService) GetAvailableTemplates(w rest.ResponseWriter, r *rest.Request) {

}

func (t *TemplateService) GetSuggestTemplates(w rest.ResponseWriter, r *rest.Request) {

}
