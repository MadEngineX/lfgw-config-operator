package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

var AppConfig = loadConfig()

type Config struct {
	ConfigMapNamespace string `yaml:"CM_NAMESPACE" env:"CM_NAMESPACE" env-default:"infra-monitoring"`
	ConfigMapName      string `yaml:"CM_NAME" env:"CM_NAME" env-default:"lfgw-config"`
	ConfigMapFilename  string `yaml:"CM_FILENAME" env:"CM_FILENAME" env-default:"acl.yaml"`
	LogLevel           string `yaml:"LOG_LEVEL" env:"LOG_LEVEL" env-default:"info"`
}

func loadConfig() *Config {
	var cfg Config
	if _, err := os.Stat("config.yaml"); os.IsNotExist(err) {
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			log.Fatalf("Unable to ReadEnv: %v", err)
		}
	} else {
		err = cleanenv.ReadConfig("config.yaml", &cfg)
		if err != nil {
			log.Fatalf("Unable to ReadConfig: %v", err)
		}
	}

	cfg.initLogger()

	return &cfg
}

func (c *Config) initLogger() {
	const defaultLogLevel = "info"

	logLevelValue, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		log.Warnf(
			"Invalid 'LOG_LEVEL': [%v], will use default LOG_LEVEL: [%v]. "+
				"Allowed values: trace, debug, info, warn, warning, error, fatal, panic", c.LogLevel, defaultLogLevel,
		)

		logLevelValue = log.InfoLevel
	}
	log.SetLevel(logLevelValue)
}
