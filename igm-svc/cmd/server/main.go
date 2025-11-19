package main

import (
	"igm-svc/internal/config"
	"igm-svc/internal/handlers"
	"igm-svc/internal/repository"
	"igm-svc/internal/server"
	"igm-svc/internal/services"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("[main] starting IGM serivce")
	cfg, err := config.Load()
	if err!=nil{
		log.Fatalf("failed to load congig :%v",err)
	}

	db,err:=repository.NewPostgresDB(cfg.DatabaseURL)
	if err!=nil{
		log.Fatalf("failed to xonnect to databse:%v",err)
	}
	log.Println("connected to postgres")
	redisClient,err:=repository.NewRedisClient(cfg.RedisURL)
	if err!=nil{
		log.Fatalf("failed to xonnect to redis:%v",err)
	}
	log.Println("connected to redis")
	issuRepo:=repository.NewIssueRepository(db)
	redisRepo:=repository.NewRedisRepository(redisClient)

	ondcClient :=services.NewOndcClient(cfg.SubscriberID,cfg.BapURI)


	serviceConfig:=&services.Config{
		SubcriberID: cfg.SubscriberID,
		BAPURI: cfg.BapURI,
	}

	issueService:= services.NewIssueService(issuRepo,redisRepo,ondcClient,serviceConfig)

	issueHandler :=handlers.NewIssueHandler(issueService)

	grpcServer:=server.NewGRPCServer(cfg.GRPCPort,issueHandler)

	go func(){
		sigChan:=make(chan os.Signal,1)
		signal.Notify(sigChan,os.Interrupt,syscall.SIGTERM)
		<-sigChan


        log.Println("\nReceived shutdown signal")
        grpcServer.Stop()
        os.Exit(0)
	}()

	log.Printf("🎯 IGM Service starting on %s", cfg.GRPCPort)
    if err := grpcServer.Start(); err != nil {
        log.Fatalf("Failed to start gRPC server: %v", err)
    }
}
