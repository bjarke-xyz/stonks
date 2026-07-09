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
	Port      int
	DbConnStr string

	JobKey string

	AppEnv    string
	BuildTime *time.Time

	YFinanceAPIAuthKey string
}

const (
	AppEnvDevelopment = "development"
	AppEnvProduction  = "production"
)

// journal_mode is set explicitly, not left to the default: it is persisted in the
// db file, so an inherited WAL database stays WAL until told otherwise. It must be
// a rollback journal because sqlite-backer-upper backs this db up over a read-only
// bind mount, and a WAL database cannot be opened without write access for its
// -shm file, even with no writers.
func (c *Config) ConnectionString() string {
	return fmt.Sprintf("file:%s?_pragma=journal_mode(delete)&_pragma=busy_timeout(5000)", c.DbConnStr)
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
		Port:               pkg.MustAtoi(os.Getenv("PORT")),
		DbConnStr:          os.Getenv("DB_CONN_STR"),
		JobKey:             os.Getenv("JOB_KEY"),
		AppEnv:             os.Getenv("APP_ENV"),
		YFinanceAPIAuthKey: os.Getenv("YFINANCEAPI_AUTH_KEY"),
		BuildTime:          buildTime,
	}, nil
}
