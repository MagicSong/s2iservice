package httphandler

import (
	"context"
	"net/http"

	"github.com/mongodb/mongo-go-driver/mongo/findopt"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang/glog"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"

	"github.com/MagicSong/s2iservice/pkg/constants"
	"github.com/MagicSong/s2iservice/pkg/models"
	"github.com/MagicSong/s2iservice/pkg/scm"
	"github.com/MagicSong/s2iservice/pkg/utils/idutils"
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

func (t *TemplateService) GetTemplatesHandler(w rest.ResponseWriter, r *rest.Request) {
	col := t.Db.Collection(constants.S2ITemplateCollectionName)
	limitOpt := findopt.Limit(10)
	cur, err := col.Find(context.Background(), nil, limitOpt)
	defer cur.Close(context.Background())
	tmplts := make([]*models.S2ITemplate, 0)
	for cur.Next(context.Background()) {
		tmplt := new(models.S2ITemplate)
		err = cur.Decode(tmplt)
		if err != nil {
			glog.Errorf("Error: %v", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmplts = append(tmplts, tmplt)
	}
	if err := cur.Err(); err != nil {
		glog.Errorf("Error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(tmplts)
}

func (t *TemplateService) AddTemplatesHandler(w rest.ResponseWriter, r *rest.Request) {
	username := r.Env["REMOTE_USER"].(string)
	if username != "admin" {
		rest.Error(w, "Unauthorized Operation", http.StatusNotFound)
		return
	}
	template := new(models.S2ITemplate)
	err := r.DecodeJsonPayload(template)
	if err != nil {
		rest.Error(w, "格式化模板出现错误，请检查JSON格式和字段是否正确", http.StatusBadRequest)
		return
	}
	template.ID = idutils.GetUuid(constants.S2ITemplatePrefix)
	col := t.Db.Collection(constants.S2ITemplateCollectionName)
	_, err = col.InsertOne(context.Background(), template)
	if err != nil {
		glog.Errorf("Error when insert new template: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (t *TemplateService) GetSuggestTemplatesHandler(w rest.ResponseWriter, r *rest.Request) {
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
	if len(m) < 1 {
		rest.Error(w, "Cannot recognize the code language", http.StatusNotFound)
		return
	}
	//推荐两种模板
	col := t.Db.Collection(constants.S2ITemplateCollectionName)
	filter := bson.NewDocument(bson.EC.ArrayFromElements("$or", bson.VC.DocumentFromElements(
		bson.EC.String("language", m[0]),
		bson.EC.String("language", m[1])),
	))
	cur, err := col.Find(context.Background(), filter)
	if err != nil {
		glog.Errorf("Error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())
	tmplts := make([]*models.S2ITemplate, 0)
	for cur.Next(context.Background()) {
		tmplt := new(models.S2ITemplate)
		err = cur.Decode(tmplt)
		if err != nil {
			glog.Errorf("Error: %v", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmplts = append(tmplts, tmplt)
	}
	if err := cur.Err(); err != nil {
		glog.Errorf("Error: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(tmplts)
}
