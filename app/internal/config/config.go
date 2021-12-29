package config

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Ifconfig struct {
		Host string `yaml:"host"`
		Uri  string `yaml:"uri"`
	} `yaml:"ifconfig"`
	DigitalOcean struct {
		Token string `yaml:"token" envconfig:"DO_TOKEN"`
	} `yaml:"digitalocean"`
	Domains  []string `yaml:"domains"`
	Interval int      `yaml:"interval"`
}

func (c *Config) Read(f string) {
	buf, err := os.Open(f)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer buf.Close()

	decoder := yaml.NewDecoder(buf)
	err = decoder.Decode(c)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func (c *Config) ReadEnv() {
	err := envconfig.Process("", c)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
