package config

import (
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/macrosiak/rspi-timelaps-manager-go/camera"
	"github.com/rs/zerolog/log"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode"
)

type Config struct {
	Development bool `default:"false" split_words:"true"`
	Streaming   bool `default:"false" split_words:"true"`

	WebInterface          bool   `default:"true" split_words:"true"`
	WebInterfaceFilesPath string `default:"./web_client" split_words:"true"`
	Password              string `default:"admin" split_words:"true"`

	OutputDir string        `default:"photos" split_words:"true"`
	Delay     time.Duration `default:"1m" split_words:"true"`

	AutoFocusRange camera.AutoFocusRange `default:"normal" split_words:"true"`
	AutoFocusMode  camera.AutoFocusMode  `default:"auto" split_words:"true"`
	Quality        int                   `default:"95" split_words:"true"`
	Hdr            bool                  `default:"false" split_words:"true"`
	VFlip          bool                  `default:"false" split_words:"true"`
	HFlip          bool                  `default:"false" split_words:"true"`
	Encoding       camera.Encoding       `default:"jpg" split_words:"true"`
	Denoise        camera.Denoise        `default:"auto" split_words:"true"`
}

func New() *Config {
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to process env")
		return nil
	}

	_ = os.Mkdir(cfg.OutputDir, 0755)
	return cfg
}

func GenerateEnvTemplate() {
	cfg := Config{}
	t := reflect.TypeOf(cfg)
	fmt.Println("Here are the expected environment variables:")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("split_words")

		if tag == "true" {
			envName := field.Name
			runes := []rune(envName)
			var parts []string
			word := []rune{}

			for i, r := range runes {
				if unicode.IsUpper(r) {
					if len(word) > 0 {
						parts = append(parts, string(word))
					}
					word = []rune{runes[i]}
				} else {
					word = append(word, r)
				}
			}
			parts = append(parts, string(word))
			envName = strings.ToUpper(strings.Join(parts, "_"))
			fmt.Println(envName + "=")
		}
	}
}
