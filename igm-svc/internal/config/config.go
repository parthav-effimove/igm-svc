package config

import (
	"fmt"
	"os"
)

type Config struct{
	DatabaseURL string
	RedisURL string
}

func Load()(*Config,error){
	cfg :=&Config{
		DatabaseURL: getEnv("DATABASE_URL",""),
		RedisURL: getEnv("REDIS_URL","localhost:6379"),
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