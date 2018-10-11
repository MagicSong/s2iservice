package server

import (
	"context"

	"github.com/MagicSong/s2iservice/pkg/config"
	"github.com/MagicSong/s2iservice/pkg/logger"
	"github.com/adjust/rmq"
	"github.com/go-redis/redis"
	"github.com/google/go-github/github"
	"github.com/mongodb/mongo-go-driver/mongo"
	"golang.org/x/oauth2"
)

type Resources struct {
	Db           *mongo.Database
	Redis        rmq.Queue
	Connection   rmq.Connection
	GithubClient *github.Client
	cfg          *config.Config
}

type Server struct {
	Resources  *Resources
	HttpServer *HttpServer
	Worker     *Worker
}

func (s *Server) LoadResource(cfg *config.Config) {
	s.Resources = new(Resources)
	s.Resources.cfg = cfg
	s.openDatabase()
	s.openRedis()
	s.HttpServer = new(HttpServer)
	s.Worker = new(Worker)
	s.Worker.MaxConsumerCountPerSecond = cfg.Redis.MaxConsumer
}

func Serve(cfg *config.Config) {
	s := new(Server)
	s.LoadResource(cfg)
	go s.Worker.Work(s.Resources)
	s.HttpServer.Serve(s.Resources)
}

func (s *Server) openDatabase() {
	client, err := mongo.NewClient(s.Resources.cfg.MongoDB.GetUrl())
	if err != nil {
		logger.Critical("failed to initialize mongodb")
		panic(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		logger.Critical("failed to connect mongodb")
		panic(err)
	}
	s.Resources.Db = client.Database(s.Resources.cfg.MongoDB.Database)
}

func (s *Server) openRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:     s.Resources.cfg.Redis.Address,
		Password: s.Resources.cfg.Redis.Password, // no password set
		DB:       s.Resources.cfg.Redis.DB,       // use default DB
	})
	connection := rmq.OpenConnectionWithRedisClient(s.Resources.cfg.Redis.RMQName, client)
	s.Resources.Connection = connection
	s.Resources.Redis = connection.OpenQueue("S2ITask")
}

func (s *Server) openGithubClient() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Resources.cfg.Github.AuthToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	s.Resources.GithubClient = client
}
