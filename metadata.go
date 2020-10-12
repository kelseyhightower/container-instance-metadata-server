package main

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"
)

type Metadata struct {
	Instance InstanceMetadata `json:"instance"`
	Project  ProjectMetadata  `json:"project"`
}

type InstanceMetadata struct {
	ID     string `json:"id"`
	Region string `json:"region"`
}

type ProjectMetadata struct {
	NumericProjectID string `json:"numeric_project_id"`
	ProjectID        string `json:"project_id"`
}

func MetadataFromFile(path string) (*Metadata, error) {
	if path == "" {
		return nil, errors.New("metadata file required.")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	if m.Instance.ID == "" {
		m.Instance.ID = generateInstanceID()
	}

	return &m, nil
}

func generateInstanceID() string {
	now := time.Now().String()
	return fmt.Sprintf("%x", sha512.Sum512([]byte(now)))
}
