package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"service_b_reference/internal/handler"
	"service_b_reference/internal/service"
	pb "warehouse_microservices/proto/compatibility"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	checker := service.NewCompatibilityChecker()
	grpcHandler := handler.NewGRPCHandler(checker)

	pb.RegisterCompatibilityServiceServer(grpcServer, grpcHandler)

	reflection.Register(grpcServer)

	log.Println("========================================")
	log.Println("Service B (Reference Service) запущен")
	log.Println("Слушаем порт: :50051")
	log.Println("========================================")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
