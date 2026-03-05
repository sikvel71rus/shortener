package flag

import (
	"flag"
)

type FlagConfig struct {
	ServerAddress string
	BaseURL       string
}

func Parse() FlagConfig {
	cfg := FlagConfig{}
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address to run HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base address for shortened URL")
	flag.Parse()
	return cfg
}
