package main

import (
	"github.com/MagicSong/s2iservice/pkg/config"
	"github.com/MagicSong/s2iservice/pkg/server"
)

func main() {
	cfg := config.LoadConf()
	server.Serve(cfg)
}
