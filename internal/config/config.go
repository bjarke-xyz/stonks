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

// Deliberately stays on the default rollback journal: sqlite-backer-upper opens
// the db read-write, which a WAL database cannot do through a read-only mount.
func (c *Config) ConnectionString() string {
	return fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)", c.DbConnStr)
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
