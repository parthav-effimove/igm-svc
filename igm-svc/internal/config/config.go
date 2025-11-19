package config

import (
	"fmt"
	"os"
)

type Config struct{
	DatabaseURL string
	RedisURL string
	GRPCPort string
	SubscriberID string
	BapURI string
	
}

func Load()(*Config,error){
	cfg :=&Config{
		DatabaseURL: getEnv("DATABASE_URL",""),
		RedisURL: getEnv("REDIS_URL","localhost:6379"),
		GRPCPort: getEnv("GRPC_PORT",":50053"),
		SubscriberID: getEnv("SUBSCRIBER_ID","preprod.effimove.in"),
		BapURI: getEnv("BAP_URI","https://preprod.effimove.in"),
		
	}
	if cfg.DatabaseURL==""{
		return nil,fmt.Errorf("DATABASE_URL is required")
	}
	
	return cfg,nil
}

func getEnv(key,defaultValue string)string{
	value:=os.Getenv(key)
	if value!=""{
		return value
	}
	return defaultValue
}