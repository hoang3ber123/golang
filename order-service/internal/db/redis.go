package db

import (
	"context"
	"fmt"
	"log"
	"order-service/config"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	RedisDB *redis.Client
	Ctx     = context.Background()
)

func InitRedis() {
	// Convert REDIS_DATABASE to int
	dbIndex, err := strconv.Atoi(config.Config.RedisIndex)
	if err != nil {
		log.Fatalf("Invalid REDIS_DATABASE value: %v", err)
	}

	// Convert REDIS_DATABASE to int
	dbProtocol, err := strconv.Atoi(config.Config.RedisProtocol)
	if err != nil {
		log.Fatalf("Invalid REDIS_DATABASE value: %v", err)
	}

	RedisDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Config.RedisHost, config.Config.RedisPort),
		Password: config.Config.RedisPassword, // No password set
		DB:       dbIndex,                     // Use default DB
		Protocol: dbProtocol,                  // Connection protocol
	})

	// Check connection
	_, err = RedisDB.Ping(Ctx).Result()
	if err != nil {
		fmt.Println("Error when connecting to Redis:", err)
	} else {
		fmt.Println("Connected to the redis successfully!")
	}
}
