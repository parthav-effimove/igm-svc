package server

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		log.Printf("grpc call:%s", info.FullMethod)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			log.Printf("← gRPC call: %s [ERROR] duration=%v err=%v",
				info.FullMethod, duration, err)
		} else {
			log.Printf("← gRPC call: %s [OK] duration=%v",
				info.FullMethod, duration)
		}
		return resp, err
	}

}

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered in %s:%v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal server erro:%v", r)
			}
		}()
		return handler(ctx, req)
	}
}
