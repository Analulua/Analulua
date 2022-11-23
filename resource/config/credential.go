package config

import (
	"bytes"
	secretManager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	secretManagerPB "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"time"
)

type (
	Credential struct {
		Service struct {
			HttpPort string `yaml:"httpPort"`
			GrpcPort string `yaml:"grpcPort"`
		} `yaml:"service"`
		Redis struct {
			Host     string `yaml:"host"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"redis"`
		Database struct {
			Host        string        `yaml:"host"`
			Port        string        `yaml:"port"`
			User        string        `yaml:"user"`
			Password    string        `yaml:"password"`
			Name        string        `yaml:"name"`
			MaxOpen     int           `yaml:"maxOpen"`
			MaxIdle     int           `yaml:"maxIdle"`
			MaxLifetime time.Duration `yaml:"maxLifetime"`
			MaxIdleTime time.Duration `yaml:"maxIdleTime"`
		} `yaml:"database"`
		PubSub struct {
			ProjectID        string `yaml:"projectID"`
			CredentialBase64 string `yaml:"credentialBase64"`
		} `yaml:"pubSub"`
		Xxl struct {
			ServerAddr   string `yaml:"serverAddr"`
			AccessToken  string `yaml:"accessToken"`
			ExecutorPort string `yaml:"executorPort"`
			ExecutorName string `yaml:"executorName"`
		} `yaml:"xxl"`
	}
)

func NewCredential(path string) (Credential, error) {
	gsmCredentialPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	parent := os.Getenv("PARENT_CREDENTIAL")
	version := os.Getenv("VERSION")

	var configRead io.Reader
	if gsmCredentialPath != "" && parent != "" && version != "" {
		configRead = getGSMReader(path, version)
	}
	if configRead == nil {
		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			return Credential{}, err
		}
		configRead = file
	}
	var credential Credential
	dec := yaml.NewDecoder(configRead)
	if err := dec.Decode(&credential); err != nil {
		return Credential{}, err
	}
	return credential, nil
}

func getGSMReader(path, version string) io.Reader {
	ctx := context.Background()
	client, err := secretManager.NewClient(ctx)

	accessRequest := secretManagerPB.AccessSecretVersionRequest{Name: path + "/" + version}
	res, err := client.AccessSecretVersion(ctx, &accessRequest)
	if err != nil {
		log.Fatal("read gsm config failed")
	}
	if len(res.Payload.Data) <= 0 {
		return nil
	}
	return bytes.NewReader(res.Payload.Data)
}
