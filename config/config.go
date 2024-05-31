package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

const (
	dbPrefix     = "MYSQL_"
	cachePrefix  = "REDIS_"
	serverPrefix = "SERVER_"
)

type DBConfig struct {
	Host     string `env:"HOST, required"`
	Port     string `env:"PORT, required"`
	User     string `env:"USER, required"`
	Password string `env:"PASSWORD, required"`
	DBName   string `env:"DB_NAME, required"`
}

type CacheConfig struct {
	Addr     string `env:"ADDR, required"`
	Password string `env:"PASSWORD, required"`
	DB       int    `env:"DB, required"`
}

type ServerConfig struct {
	ReadTimeout               time.Duration `env:"READ_TIMEOUT,default=5s"`
	WriteTimeout              time.Duration `env:"WRITE_TIMEOUT,default=10s"`
	IdleTimeout               time.Duration `env:"IDLE_TIMEOUT,default=15s"`
	GracefulShutdownTimeout   time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT,default=5s"`
	PreflightCacheDurationSec int           `env:"PREFLIGHT_CACHE_DURATION_SEC,default=300"`
}

func NewDBConfig(ctx context.Context) (*DBConfig, error) {
	conf := &DBConfig{}
	pl := envconfig.PrefixLookuper(dbPrefix, envconfig.OsLookuper())
	if err := envconfig.ProcessWith(ctx, conf, pl); err != nil {
		return nil, err
	}
	return conf, nil
}

func NewCacheConfig(ctx context.Context) (*CacheConfig, error) {
	conf := &CacheConfig{}
	pl := envconfig.PrefixLookuper(cachePrefix, envconfig.OsLookuper())
	if err := envconfig.ProcessWith(ctx, conf, pl); err != nil {
		return nil, err
	}
	return conf, nil
}

func NewServerConfig(ctx context.Context) (*ServerConfig, error) {
	conf := &ServerConfig{}
	pl := envconfig.PrefixLookuper(serverPrefix, envconfig.OsLookuper())
	if err := envconfig.ProcessWith(ctx, conf, pl); err != nil {
		return nil, err
	}
	return conf, nil
}
