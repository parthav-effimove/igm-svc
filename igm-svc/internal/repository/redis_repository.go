package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)


type RedisRepository interface{
	SaveIssueResponse(ctx context.Context, transactionID string, payload map[string]interface{})error
	GetIssueResponse(ctx context.Context,transactionID string)([]map[string]interface{},error)
	Exists(ctx context.Context,transactionID string)(bool,error)
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
    if err!=nil{
        return fmt.Errorf("failed to marshal the payload:%w",err)
    }

    if err := r.client.RPush(ctx,key,jsonData).Err(); err != nil {
        return fmt.Errorf("failed to push to redis :%w",err)
    }

    r.client.Expire(ctx,key,r.ttl)

    return nil
}

func (r *redisRepository) GetIssueResponse(ctx context.Context,transactionID string)([]map[string]interface{},error){
	key :=fmt.Sprintf("on_issue:%s",transactionID)
	
	result,err :=r.client.LRange(ctx,key,0,-1).Result()
	if err!=nil{
		if err==redis.Nil{
			return []map[string]interface{}{},nil
		}
		return nil,fmt.Errorf("redis LRANGE failed:%w",err)
	}

	var response []map[string]interface{}
	for _,jsonStr :=range result{
		var payload map[string]interface{}
		err :=json.Unmarshal([]byte(jsonStr),&payload)
		if err!=nil{
			continue
		}
		response=append(response, payload)
	}
	return  response,nil
}

func (r *redisRepository)Exists(ctx context.Context,transactionID string)(bool,error){
	key:=fmt.Sprintf("on_issue:%s",transactionID)
	count,err :=r.client.Exists(ctx,key).Result()
	if err!=nil{
		return false,err
	}
	return  count>0,nil
}