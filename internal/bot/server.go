package bot

import (
	"errors"
	"fmt"
	"github.com/VATUSA/discord-bot-v3/internal/config"
	"github.com/VATUSA/discord-bot-v3/pkg/constants"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

type ServerConfig struct {
	ID             string                   `yaml:"id"`
	Name           string                   `yaml:"name"`
	Active         bool                     `yaml:"active"`
	Description    string                   `yaml:"description"`
	Facility       string                   `yaml:"facility"`
	NameFormatType constants.NameFormatType `yaml:"name_format_type"`
	TitleType      constants.TitleType      `yaml:"title_type"`
	Roles          []RoleConfig             `yaml:"roles"`
}

type RoleConfig struct {
	ID       string           `yaml:"id"`
	Name     string           `yaml:"name"`
	Criteria []CriteriaConfig `yaml:"criteria"`
}

type CriteriaConfig struct {
	Name       string            `yaml:"name"`
	Conditions []ConditionConfig `yaml:"conditions"`
}

type ConditionConfig struct {
	Type   constants.ConditionType `yaml:"type"`
	Value  *string                 `yaml:"value"`
	Invert bool                    `yaml:"invert" default:"false"`
}

func LoadAllServerConfigOrPanic(configPath string) map[string]ServerConfig {
	configs, err := LoadAllServerConfig(configPath)
	if err != nil {
		log.Printf(err.Error())
	}
	return configs
}

func LoadAllServerConfig(configPath string) (map[string]ServerConfig, error) {
	configs := make(map[string]ServerConfig, 0)
	files, err := os.ReadDir(configPath)
	if err != nil {
		return nil, errors.New("failed to load server configs")
	}
	for _, f := range files {
		if !f.IsDir() {
			cfg, err := LoadServerConfig(fmt.Sprintf("%s/%s", configPath, f.Name()))
			if err != nil {
				log.Printf(err.Error())
				return nil, nil
			}
			configs[cfg.ID] = *cfg
		}
	}
	return configs, nil
}

func LoadServerConfig(configPath string) (*ServerConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg ServerConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	// TODO: Validate that roles aren't duplicated
	// TODO: Validate role criteria
	log.Printf("Loaded Config for %s (%s)\n", cfg.Name, cfg.ID)
	return &cfg, nil
}

var configs = LoadAllServerConfigOrPanic(config.CONFIG_PATH)

func GetServerConfig(id string) *ServerConfig {
	cfg, ok := configs[id]
	if !ok {
		return nil
	}
	return &cfg
}
