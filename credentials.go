package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type Credentials struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	QuotaProjectID string `json:"quota_project_id"`
	RefreshToken   string `json:"refresh_token"`
	Type           string `json:"type"`
}

type ServiceAccount struct {
	Type         string `json:"type"`
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURL     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`
}

func CredentialsFromFile() (*Credentials, error) {
	f := wellKnownFile()
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("Unable to read credentials file: %w", err)
	}

	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("Unable to parse credentials file: %w", err)
	}

	return &c, nil
}

func ServiceAccountFromFile(path string) (*ServiceAccount, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read service account file: %w", err)
	}

	var sa ServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return nil, fmt.Errorf("Unable to parse service account file: %w", err)
	}

	return &sa, nil
}

func wellKnownFile() string {
	if defaultCredentials != "" {
		return defaultCredentials
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		if u, err := user.Current(); err == nil {
			homeDir = u.HomeDir
		}
	}
	return filepath.Join(homeDir, ".config", "gcloud", "application_default_credentials.json")
}
