package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(redisURL string)(*redis.Client,error){
	client:=redis.NewClient(&redis.Options{
		Addr: redisURL,
		Password: "",
		DB: 0,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize: 10,

	})

	context,cancel :=context.WithTimeout(context.Background(),5*time.Second)

	defer cancel()

	err := client.Ping(context).Err()
	if err !=nil{
		return nil,fmt.Errorf("failed to connect to redis:%w",err)
	}

	log.Println("connected to redis")
	return  client,nil
}