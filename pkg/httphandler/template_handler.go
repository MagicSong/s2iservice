package httphandler

import (
	"net/http"

	"github.com/MagicSong/s2iservice/pkg/scm"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type TemplateService struct {
	token string
	Db    *mongo.Database
}

func NewTemplateService(db *mongo.Database, token string) *TemplateService {
	return &TemplateService{
		Db:    db,
		token: token,
	}
}

func (t *TemplateService) GetAvailableTemplates(w rest.ResponseWriter, r *rest.Request) {

}

func (t *TemplateService) GetSuggestTemplates(w rest.ResponseWriter, r *rest.Request) {
	v := r.URL.Query()
	s := v.Get("source")
	if s == "" {
		rest.Error(w, "source is required", http.StatusBadRequest)
		return
	}
	git, err := scm.NewTokenClient(t.token, s, 0)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m, err := git.ListLanguages()
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(m)
}
