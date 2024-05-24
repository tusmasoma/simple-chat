package config

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewClient() *redis.Client {
	ctx := context.Background()
	conf, err := NewCacheConfig(ctx)
	if err != nil || conf == nil {
		log.Printf("Failed to load cache config: %s\n", err)
		return nil
	}

	client := redis.NewClient(&redis.Options{Addr: conf.Addr, Password: conf.Password, DB: conf.DB})
	return client
}
