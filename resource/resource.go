package resource

import (
	"newdemo1/resource/config"
	"newdemo1/resource/datadog"
	"newdemo1/resource/jaeger"
	"newdemo1/resource/logger"
	"newdemo1/resource/validator"
)

type Resource struct {
	Config     config.Configuration
	Credential config.Credential
	Jaeger     jaeger.Jaeger
	Log        logger.Logger
	Datadog    datadog.DataDog
	Validator  validator.Validator
}

func NewResource(configurationPath, credentialPath string) (*Resource, error) {

	//导入配置文件
	configuration, err := config.NewConfiguration(configurationPath)
	if err != nil {
		return nil, err
	}

	//导入数据库文件
	credential, err := config.NewCredential(credentialPath)
	if err != nil {
		return nil, err
	}
	jaeger, err := jaeger.NewJaeger(configuration)
	if err != nil {
		return nil, err
	}
	telemetryLogger, err := logger.NewLogger(jaeger.Tracer)
	if err != nil {
		return nil, err
	}
	dataDog, err := datadog.NewDataDog(configuration, jaeger.Tracer)
	if err != nil {
		return nil, err
	}
	dataValidator, err := validator.NewValidator()
	if err != nil {
		return nil, err
	}

	return &Resource{
		Config:     configuration,
		Credential: credential,
		Jaeger:     jaeger,
		Log:        telemetryLogger,
		Datadog:    dataDog,
		Validator:  dataValidator,
	}, nil
}

func (r *Resource) Flush() {
	r.Jaeger.FlushFn()
}
