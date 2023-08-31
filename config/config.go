package config

import (
	"encoding/json"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"os"
	"time"
)

type Config struct {
	Development bool          `default:"false" split_words:"true"`
	OutputDir   string        `default:"photos" split_words:"true"`
	Delay       time.Duration `default:"1s" split_words:"true"`

	AutoFocusRange camera.AutoFocusRange `default:"normal" split_words:"true"`
	Quality        int                   `default:"95" split_words:"true"`
	HDR            bool                  `default:"false" split_words:"true"`
	VFlip          bool                  `default:"false" split_words:"true"`
	HFlip          bool                  `default:"false" split_words:"true"`
	Encoding       camera.Encoding       `default:"jpg" split_words:"true"`
}

var cfg *Config

func New() *Config {
	if cfg != nil {
		return cfg
	}
	cfg = &Config{}

	err := envconfig.Process("", cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func Save() error {
	by, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	file, err := os.OpenFile("config.cfg", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = file.Write(by)
	return err
}
