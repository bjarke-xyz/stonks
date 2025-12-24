package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bjarke-xyz/stonks/pkg"
	"github.com/joho/godotenv"
)

type Config struct {
	Port             int
	DbConnStr        string
	BackupDbPath     string
	TursoDatabaseUrl string
	TursoAuthToken   string

	S3BackupUrl             string
	S3BackupBucket          string
	S3BackupAccessKeyId     string
	S3BackupSecretAccessKey string

	JobKey string

	AppEnv    string
	BuildTime *time.Time

	CookieSecret string

	YFinanceAPIAuthKey string
}

const (
	AppEnvDevelopment = "development"
	AppEnvProduction  = "production"
)

func (c *Config) ConnectionString() string {
	// return c.DbConnStr
	return fmt.Sprintf("%s?authToken=%s", c.TursoDatabaseUrl, c.TursoAuthToken)
}

func NewConfig() (*Config, error) {
	godotenv.Load()
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = AppEnvDevelopment
	} else {
		if appEnv != AppEnvDevelopment && appEnv != AppEnvProduction {
			return nil, fmt.Errorf("failed to validate APP_ENV: invalid value %q", appEnv)
		}
	}
	buildTimeStr := os.Getenv("BUILD_TIME")
	var buildTime *time.Time
	if buildTimeStr != "" {
		_buildTime, err := time.Parse("2006-01-02 15:04:05", buildTimeStr)
		if err != nil {
			log.Printf("error parsing BUILD_TIME env: %v", err)
		}
		buildTime = &_buildTime
	}
	return &Config{
		Port:                    pkg.MustAtoi(os.Getenv("PORT")),
		DbConnStr:               os.Getenv("DB_CONN_STR"),
		BackupDbPath:            os.Getenv("BACKUP_DB_PATH"),
		TursoDatabaseUrl:        os.Getenv("TURSO_DATABASE_URL"),
		TursoAuthToken:          os.Getenv("TURSO_AUTH_TOKEN"),
		S3BackupUrl:             os.Getenv("S3_BACKUP_URL"),
		S3BackupBucket:          os.Getenv("S3_BACKUP_BUCKET"),
		S3BackupAccessKeyId:     os.Getenv("S3_BACKUP_ACCESS_KEY_ID"),
		S3BackupSecretAccessKey: os.Getenv("S3_BACKUP_SECRET_ACCESS_KEY"),
		JobKey:                  os.Getenv("JOB_KEY"),
		AppEnv:                  os.Getenv("APP_ENV"),
		CookieSecret:            os.Getenv("COOKIE_SECRET"),
		YFinanceAPIAuthKey:      os.Getenv("YFINANCEAPI_AUTH_KEY"),
		BuildTime:               buildTime,
	}, nil
}
