package config

import (
	"bytes"
	secretManager "cloud.google.com/go/secretmanager/apiv1"
	src "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
)

type (
	Configuration struct {
		Service struct {
			Name     string `yaml:"name"`
			HttpPort string `yaml:"httpPort"`
			GrpcPort string `yaml:"grpcPort"`
		} `yaml:"service"`
		Telemetry struct {
			Tracer struct {
				CollectorEndpoint string `yaml:"collectorEndpoint"`
				ServiceName       string `yaml:"serviceName"`
				SourceEnv         string `yaml:"sourceEnv"`
			} `yaml:"tracer"`
			Metric struct {
				Port         int     `yaml:"port"`
				AgentAddress string  `yaml:"agentAddress"`
				SampleRate   float64 `yaml:"sampleRate"`
				DatadogKey   string  `yaml:"datadogKey"`
			} `yaml:"metric"`
			Filter struct {
				Body   []string `yaml:"body"`
				Header []string `yaml:"header"`
			}
		} `yaml:"telemetry"`
		Pubsub struct {
			PublishTopic struct {
				RecurringHappen string `yaml:"recurring-happen"`
			} `yaml:"publishTopic"`
			Subscriber struct {
				SubscriptionHappenResult string `yaml:"subscriptionHappenResult"`
				SubscriptionJobFinish    string `yaml:"subscriptionJobFinish"`
			} `yaml:"subscriber"`
		} `yaml:"pubSub"`
	}
)

// NewConfiguration 读取配置
func NewConfiguration(path string) (Configuration, error) {
	//读取环境变量
	gsmCredentialPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	parent := os.Getenv("PARENT")
	version := os.Getenv("VERSION")
	var configRead io.Reader
	if "" != gsmCredentialPath && "" != parent && version != "" {
		configRead = gegGSMReader(parent, version)
	}

	if configRead == nil {
		file, err := os.Open(path)
		defer file.Close()
		if err != nil {
			return Configuration{}, err
		}
		configRead = file
	}

	dec := yaml.NewDecoder(configRead)
	var config Configuration
	err := dec.Decode(&config)
	if err != nil {
		return Configuration{}, err
	}
	return config, nil
}
func gegGSMReader(parent, version string) io.Reader {
	ctx := context.Background()
	client, err := secretManager.NewClient(ctx)
	//构建请求
	accessRequest := src.AccessSecretVersionRequest{Name: parent + "/" + version}
	//请求API
	result, err := client.AccessSecretVersion(ctx, &accessRequest)
	if err != nil {
		log.Fatal("read gsm config failed")
	}
	if len(result.Payload.Data) <= 0 {
		return nil
	}

	return bytes.NewReader(result.Payload.Data)
}
