package main

import (
	"context"
	"fmt"
	gRPCRestaurant "gRPCProcessingServer/restaurantService"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Реализация сервера, которая соответствует интерфейсу OrderCreaterServer
type server struct {
	gRPCRestaurant.UnimplementedProcessCreaterServer
}

// Реализация метода Create
func (s *server) Create(ctx context.Context, req *gRPCRestaurant.OrderDetails) (*gRPCRestaurant.OrderStatus, error) {
	fmt.Println("Заказ с id %d получен", req.OrderID)
	for _, elem := range req.Dishes {
		fmt.Println(elem)
	}
	return &gRPCRestaurant.OrderStatus{OrderID: 1, Status: "в готовке"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Создаем новый gRPC-сервер
	grpcServer := grpc.NewServer()

	// Регистрируем наш Calculator-сервис на gRPC-сервере
	gRPCRestaurant.RegisterProcessCreaterServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	log.Println("gRPC сервер запущен на порту :50050")
	// Запускаем сервер
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
