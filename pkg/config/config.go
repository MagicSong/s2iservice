package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/MagicSong/s2iservice/pkg/logger"
	"github.com/koding/multiconfig"
)

const (
	EtcdPrefix = "devopsbuilder"
)

type Config struct {
	Log     LogConfig
	MongoDB MongoConfig
	Redis   RedisConfig
	Github  GithubConfig
}

type LogConfig struct {
	Level string `default:"info"` // debug, info, warn, error, fatal
}

type MongoConfig struct {
	Host     string `default:"192.168.98.8"`
	Port     string `default:"27017"`
	User     string `default:"root"`
	Password string `default:"magicsong"`
	Database string `default:"devops"`
}
type RedisConfig struct {
	Address     string `default:"192.168.98.3:6379"`
	Password    string
	DB          int
	RMQName     string `default:"Redis Service"`
	MaxConsumer int    `default:"2"`
}

type GithubConfig struct {
	AuthToken string `default:"5f87b50d0d77f5825dbc05926fcef9ffe8815a66"`
}

func (m *MongoConfig) GetUrl() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s", m.User, m.Password, m.Host, m.Port)
}

func PrintUsage() {
	flag.PrintDefaults()
	fmt.Fprint(os.Stdout, "\nSupported environment variables:\n")
	e := newLoader("devopsphere")
	e.PrintEnvs(new(Config))
	fmt.Println("")
}

func GetFlagSet() *flag.FlagSet {
	flag.CommandLine.Usage = PrintUsage
	return flag.CommandLine
}

func ParseFlag() {
	GetFlagSet().Parse(os.Args[1:])
}

var profilingServerStarted = false

func LoadConf() *Config {
	ParseFlag()

	config := new(Config)
	m := &multiconfig.DefaultLoader{}
	m.Loader = multiconfig.MultiLoader(newLoader("devopsphere"))
	m.Validator = multiconfig.MultiValidator(
		&multiconfig.RequiredValidator{},
	)
	err := m.Load(config)
	if err != nil {
		panic(err)
	}
	logger.SetLevelByString(config.Log.Level)
	logger.Info("LoadConf: %+v", config)

	return config
}
