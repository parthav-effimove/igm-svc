package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)


type RedisRepository interface{
	SaveIssu
}

type redisRepository struct{
	client *redis.Client
	ttl time.Duration
}

func NewRedisRepository(client *redis.Client)RedisRepository{
	return &redisRepository{
		client: client,
		ttl: 24 * time.Hour,
	}
}

func (r *redisRepository) SaveIssueResponse(ctx context.Context, transactionID string, payload map[string]interface{})error{
	key :=fmt.Sprintf("on_issue:%s",transactionID)

	jsonData,err :=json.Marshal(payload)
}