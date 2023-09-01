package config

import (
	"encoding/json"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type Config struct {
	ConfigFilePath string `json:"config_file_path"`

	Development bool          `json:"development"`
	OutputDir   string        `json:"output_dir"`
	Delay       time.Duration `json:"delay"`

	AutoFocusRange camera.AutoFocusRange `json:"auto_focus_range"`
	Quality        int                   `json:"quality"`
	HDR            bool                  `json:"hdr"`
	VFlip          bool                  `json:"vflip"`
	HFlip          bool                  `json:"hflip"`
	Encoding       camera.Encoding       `json:"encoding"`
}

func (c *Config) loadDefault() {
	c.ConfigFilePath = "config.json"
	c.Development = false
	c.OutputDir = "photos"
	c.Delay = time.Minute
	c.AutoFocusRange = "normal"
	c.Quality = 95
	c.HDR = false
	c.VFlip = false
	c.HFlip = false
	c.Encoding = "jpg"
}

func (c *Config) loadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}

	return nil
}

func New(configPath string) (*Config, error) {
	config := &Config{}

	println(configPath)
	log.Info().Msg(configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config.loadDefault()
		return config, nil
	} else {
		err := config.loadFromFile(configPath)
		if err != nil {
			return nil, err
		}
		return config, nil
	}
}
