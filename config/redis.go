package config

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var Redis *redis.Client

func NewClient() {
	ctx := context.Background()
	conf, err := NewCacheConfig(ctx)
	if err != nil || conf == nil {
		log.Printf("Failed to load cache config: %s\n", err)
		//return nil
	}

	client := redis.NewClient(&redis.Options{Addr: conf.Addr, Password: conf.Password, DB: conf.DB})
	Redis = client
	//return client
}
