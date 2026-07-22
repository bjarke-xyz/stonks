package config

import (
	"fmt"
	"log/slog"
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

func (c *Config) ConnectionString() string {
	// _time_format=sqlite makes the driver write time.Time as
	// "2006-01-02 15:04:05.999999999-07:00". Without it the driver defaults to
	// time.Time.String(), which for an unnamed fixed zone renders the offset in
	// the zone-name slot ("... +0200 +0200") — a value neither the driver nor
	// SQLite's date functions can parse back.
	return fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_time_format=sqlite", c.DbConnStr)
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
			slog.Warn("parsing BUILD_TIME env failed", "value", buildTimeStr, "error", err)
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
