package client

import (
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"newdemo1/resource"
	"newdemo1/resource/jaeger/common/mysql"
	"os"
	"time"
)

type Client struct {
	db *gorm.DB
}

func NewClient(resource *resource.Resource) (*Client, error) {
	db, err := mysql.DB(mysql.Config{
		Host:        resource.Credential.Database.Host,
		Port:        resource.Credential.Database.Port,
		User:        resource.Credential.Database.User,
		Password:    resource.Credential.Database.Password,
		Name:        resource.Credential.Database.Name,
		MaxOpen:     resource.Credential.Database.MaxOpen,
		MaxIdle:     resource.Credential.Database.MaxIdle,
		MaxLifetime: int(resource.Credential.Database.MaxLifetime.Minutes()),
		MaxIdleTime: int(resource.Credential.Database.MaxIdleTime.Minutes()),
		ParseTime:   true,
		Location:    "Asia/Jakarta",
	})
	if err != nil {
		return &Client{}, err
	}

	gormDB, err := gorm.Open(gormMysql.New(gormMysql.Config{
		Conn: db,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)},
	)
	return &Client{db: gormDB}, err
}
