package repository

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(databaseURL string)(*gorm.DB,error){
	if databaseURL==""{
		return  nil,fmt.Errorf("data base url is empty")
	}

	db,err :=gorm.Open(postgres.Open(databaseURL),&gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc:func() time.Time{
			return  time.Now().UTC()
		},
	})
	if err!=nil{
		return nil,fmt.Errorf("failed to connect to DB:%w",err)
	}

	sqlDB, err :=db.DB()
	if err!=nil{
		return  nil,fmt.Errorf("failed to get databse instance :%w",err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err=sqlDB.Ping()
	if err !=nil{
		return nil,fmt.Errorf("failed to ping database :%w",err)
	}
	log.Printf("connected to postgreSQL")
	return db,nil
}