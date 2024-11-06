package main

import (
	"context"
	"fmt"
	gRPCOrder "gRPCProcessingServer/orderService"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Реализация сервера, которая соответствует интерфейсу OrderCreaterServer
type server struct {
	gRPCOrder.UnimplementedOrderCreaterServer
}

// Реализация метода Create
func (s *server) Create(ctx context.Context, req *gRPCOrder.OrderRequest) (*gRPCOrder.OrderResponse, error) {
	for _, elem := range req.Dishes {
		fmt.Println(elem)
	}
	return &gRPCOrder.OrderResponse{OrderID: 1}, nil
}

func main() {
	// Настраиваем сервер для прослушивания порта 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаем новый gRPC-сервер
	grpcServer := grpc.NewServer()

	// Регистрируем наш Calculator-сервис на gRPC-сервере
	gRPCOrder.RegisterOrderCreaterServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	log.Println("gRPC сервер запущен на порту :50051")
	// Запускаем сервер
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
