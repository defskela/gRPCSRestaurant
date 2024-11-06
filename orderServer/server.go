package main

import (
	"context"
	"fmt"
	gRPCOrder "gRPCProcessingServer/orderService"
	gRPCRestaurant "gRPCProcessingServer/restaurantService"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Функция для создания клиента и подключения к RestaurantService
func connectToRestaurantService() (gRPCRestaurant.ProcessCreaterClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Не удалось подключиться к RestaurantService: %v", err)
	}

	client := gRPCRestaurant.NewProcessCreaterClient(conn)
	return client, conn
}

// Реализация сервера, которая соответствует интерфейсу OrderCreaterServer
type server struct {
	gRPCOrder.UnimplementedOrderCreaterServer
}

// Реализация метода Create
func (s *server) Create(ctx context.Context, req *gRPCOrder.OrderRequest) (*gRPCOrder.OrderResponse, error) {
	for _, elem := range req.Dishes {
		fmt.Println(elem)
	}
	// Создаем клиента для второго микросервиса (RestaurantService)
	restaurantClient, conn := connectToRestaurantService()
	defer conn.Close()

	// Формируем запрос к RestaurantService
	restaurantReq := &gRPCRestaurant.OrderDetails{
		OrderID: 1,
		Dishes:  req.Dishes,
	}

	// Отправляем запрос
	ctxRestaurant, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	response, err := restaurantClient.Create(ctxRestaurant, restaurantReq)
	if err != nil {
		log.Fatalf("Ошибка при вызове ProcessOrder: %v", err)
	}

	log.Printf("Получен статус от RestaurantService: %s", response.Status)

	return &gRPCOrder.OrderResponse{OrderID: 1}, nil
}

func main() {
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
