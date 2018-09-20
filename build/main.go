package main

import (
	"github.com/s2iservice/pkg/config"
	"github.com/s2iservice/pkg/server"
)

func main() {
	cfg := config.LoadConf()
	server.Serve(cfg)
}
