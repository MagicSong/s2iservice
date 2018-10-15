package httphandler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"

	"github.com/MagicSong/s2iservice/pkg/constants"
	"github.com/MagicSong/s2iservice/pkg/models"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type LogService struct {
	Db *mongo.Database
}

func NewLogService(db *mongo.Database) *LogService {
	return &LogService{
		Db: db,
	}
}

func (l *LogService) GetLoggerHandler(w rest.ResponseWriter, r *rest.Request) {
	jid := r.PathParams["jid"]
	username := r.Env["REMOTE_USER"].(string)
	coll := l.Db.Collection(constants.S2IJobCollectionName)
	filter := bson.NewDocument(bson.EC.String("_id", jid), bson.EC.String("username", username))
	projection := findopt.Projection(bson.NewDocument(bson.EC.Int32("_id", 1)))
	var t struct {
		ID string `bson:"_id"`
	}
	err := coll.FindOne(context.Background(), filter, projection).Decode(&t)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	coll = l.Db.Collection(constants.S2ILogCollectionName)
	var fromID int
	var startTime time.Time
	var retryID uint8
	req := r.URL.Query()
	s := req.Get("retry_id")
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			rest.Error(w, "retry_id不是合法的数字", http.StatusBadRequest)
			return
		}
		retryID = uint8(i)
	}

	s = req.Get("start_time")
	if s != "" {
		startTime, err = time.Parse(time.RFC3339, s)
		if err != nil {
			rest.Error(w, "start_time格式不正确", http.StatusBadRequest)
			return
		}
		res, err := getLogByTime(coll, jid, startTime, retryID)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteJson(&res)
		return

	}
	s = req.Get("from_id")
	if s != "" {
		fromID, err = strconv.Atoi(s)
		if err != nil {
			rest.Error(w, "from_id不是合法的数字", http.StatusBadRequest)
			return
		}
	}

	res, err := getLogByID(coll, fromID, jid, retryID)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteJson(&res)
}

func getLogByID(coll *mongo.Collection, fromid int, jobid string, retryid uint8) ([]*models.LogRow, error) {
	res := make([]*models.LogRow, 0)
	filter := bson.NewDocument(bson.EC.String("builder_id", jobid), bson.EC.Int32("retry_id", int32(retryid)), bson.EC.SubDocumentFromElements("seq", bson.EC.Int32("$gte", int32(fromid))))
	cur, err := coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		log := &models.LogRow{}
		err = cur.Decode(log)
		if err != nil {
			return nil, err
		}
		res = append(res, log)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func getLogByTime(coll *mongo.Collection, jobid string, startTime time.Time, retryid uint8) ([]*models.LogRow, error) {
	res := make([]*models.LogRow, 0)
	filter := bson.NewDocument(bson.EC.String("builder_id", jobid), bson.EC.Int32("retry_id", int32(retryid)), bson.EC.SubDocumentFromElements("create_time", bson.EC.Time("$gte", startTime)))
	cur, err := coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		log := &models.LogRow{}
		err = cur.Decode(log)
		if err != nil {
			return nil, err
		}
		res = append(res, log)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
