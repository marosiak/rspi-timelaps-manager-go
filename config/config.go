package config

import (
	"encoding/json"
	"flag"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	_ "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type Config struct {
	ConfigFilePath string `default:"config.json" split_words:"true"`

	Development bool          `default:"false" split_words:"true"`
	OutputDir   string        `default:"photos" split_words:"true"`
	Delay       time.Duration `default:"1m" split_words:"true"`

	AutoFocusRange camera.AutoFocusRange `default:"normal" split_words:"true"`
	Quality        int                   `default:"95" split_words:"true"`
	HDR            bool                  `default:"false" split_words:"true"`
	VFlip          bool                  `default:"false" split_words:"true"`
	HFlip          bool                  `default:"false" split_words:"true"`
	Encoding       camera.Encoding       `default:"jpg" split_words:"true"`
}

var cfg *Config

func saveFile() error {
	by, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.ConfigFilePath, by, 0644)
	if err != nil {
		return err
	}
	log.Debug().Msg("Config file saved")
	return nil
}

func loadFile() error {
	fileBytes, err := os.ReadFile(cfg.ConfigFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileBytes, cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse config file")
		return err
	}
	log.Debug().Msg("Config file loaded")
	return nil
}

func New() *Config {
	if cfg != nil {
		return cfg
	}
	cfg = &Config{}

	var tmpConfigFilePath string
	flag.StringVar(&tmpConfigFilePath, "config", "", "path to config file")
	flag.Parse()

	if tmpConfigFilePath != "" {
		cfg.ConfigFilePath = tmpConfigFilePath
	}

	err := loadFile()
	if err != nil {
		err = envconfig.Process("", cfg)
	}
	err = saveFile()
	if err != nil {
		log.Err(err).Str("configFilePath", cfg.ConfigFilePath).Msg("Failed to save config file")
		return cfg
	}
	return cfg
}

func Save() error {
	by, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	file, err := os.OpenFile("config.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = file.Write(by)
	return err
}
