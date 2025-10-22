package redis

import (
	"context"
	"fmt"
	"sample/internal/app/config"
	"time"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "recovery_service."
const jwtPrefix = "jwt."

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	// Используем cfg для создания подключения
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		//Password: cfg.Password,
		Username: cfg.User,
		DB:       0,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}
	return &Client{client: redisClient}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

type Client struct {
	client *redis.Client
}

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	return c.client.Set(ctx, getJWTKey(jwtStr), true, jwtTTL).Err()
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	return c.client.Get(ctx, getJWTKey(jwtStr)).Err()
}
