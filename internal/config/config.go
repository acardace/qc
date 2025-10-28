package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Jira struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"` // Personal Access Token
	} `yaml:"jira"`
	GitHub struct {
		Token string `yaml:"token"` // Personal Access Token
	} `yaml:"github"`
	Associates map[string]AssociateInfo `yaml:"associates"`
}

type AssociateInfo struct {
	JiraUsername   string `yaml:"jira_username"`
	GitHubUsername string `yaml:"github_username"`
	FullName       string `yaml:"full_name"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
