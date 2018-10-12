package httphandler

import (
	"net/http"

	"github.com/MagicSong/s2iservice/pkg/scm"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type TemplateService struct {
	githubToken string
	Db          *mongo.Database
}

func NewTemplateService(db *mongo.Database, token string) *TemplateService {
	return &TemplateService{
		Db:          db,
		githubToken: token,
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

	var gitType scm.GitType
	var token string
	ty := v.Get("type")

	switch ty {
	case "gitlab":
		gitType = scm.Gitlab
		token = v.Get("token")
		if token == "" {
			rest.Error(w, "Gitlab project should provide access token", http.StatusBadRequest)
			return
		}
	case "github", "":
		gitType = scm.Github
		token = t.githubToken
	default:
		rest.Error(w, "Unsupported source type", http.StatusBadRequest)
		return
	}

	git, err := scm.NewTokenClient(token, s, gitType)
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
