package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/mbertschler/foundation"
	"github.com/mbertschler/foundation/service"
	"github.com/pkg/errors"
)

// Default configuration values
const (
	defaultConfigPath = "foundation_config.json"
	defaultHostPort   = "localhost:3000"
)

// Command-line flag variables
var (
	configPath string
	devMode    bool

	defaultConfig = foundation.Config{
		HostPort: defaultHostPort,
	}
)

func main() {
	startup := time.Now()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Foundation demo server 🚀")

	flag.StringVar(&configPath, "config", defaultConfigPath, "foundation config JSON file path")
	flag.BoolVar(&devMode, "dev", false, "dev mode: serve asset files from browser/dist directory instead of Go embedded assets")
	// more flags if needed
	flag.Parse()

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatal("failed to load config:", err)
		return
	}

	err = postProcessConfig(config, startup)
	if err != nil {
		log.Fatal("failed to process config:", err)
		return
	}

	exitCode := service.RunApp(config)
	os.Exit(exitCode)
}

func loadConfig(path string) (*foundation.Config, error) {
	config := defaultConfig

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "can't read config file")
	}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Fatal("config file JSON error:", err)
	}

	return &config, nil
}

func postProcessConfig(config *foundation.Config, startup time.Time) error {
	config.Startup = startup

	if devMode {
		log.Println("dev mode, serving asset files from disk")
		config.DevFileServer = true
	}
	return nil
}
