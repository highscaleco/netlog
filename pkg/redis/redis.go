package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func redisDBFromEnv() int {
	dbStr := os.Getenv("REDIS_DB")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return 0 // Default to DB 0 if conversion fails
	}
	return db
}

func createClient() *redis.Client {
	if rdb != nil {
		return rdb
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDBFromEnv(),
	})
	return rdb
}

type IPInfo struct {
	Namespace string
	Name      string
}

func SetIP(ip string, info IPInfo) error {
	if ip == "" {
		return fmt.Errorf("ip cannot be empty")
	}

	client := createClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convert struct to map for HMSet
	fields := map[string]interface{}{
		"namespace": info.Namespace,
		"name":      info.Name,
	}

	_, err := client.HMSet(ctx, ip, fields).Result()
	if err != nil {
		return fmt.Errorf("failed to set IP info: %w", err)
	}
	return nil
}

func GetIP(ip string) (IPInfo, error) {
	if ip == "" {
		return IPInfo{}, fmt.Errorf("ip cannot be empty")
	}

	client := createClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := client.HGetAll(ctx, ip).Result()
	if err != nil {
		return IPInfo{}, fmt.Errorf("failed to get IP info: %w", err)
	}

	// Check if the key exists
	if len(info) == 0 {
		return IPInfo{}, fmt.Errorf("no info found for ip: %s", ip)
	}

	// Check if required fields exist
	namespace, ok := info["namespace"]
	if !ok {
		return IPInfo{}, fmt.Errorf("namespace field missing for ip: %s", ip)
	}

	name, ok := info["name"]
	if !ok {
		return IPInfo{}, fmt.Errorf("name field missing for ip: %s", ip)
	}

	return IPInfo{
		Namespace: namespace,
		Name:      name,
	}, nil
}
