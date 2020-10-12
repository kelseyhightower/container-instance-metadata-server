package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type ServiceAccount struct {
	Type         string `json:"type"`
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURL     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`
}

func ServiceAccountFromFile(path string) (*ServiceAccount, error) {
	data, err := ioutil.ReadFile(serviceAccountFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read service account file: %w", err)
	}

	var sa ServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return nil, fmt.Errorf("Unable to parse service account file: %w", err)
	}

	return &sa, nil
}
