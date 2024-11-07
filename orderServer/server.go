package main

import (
	"context"
	"fmt"
	gRPCOrder "gRPCProcessingServer/orderService"
	gRPCRestaurant "gRPCProcessingServer/restaurantService"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type OrderDetails struct {
	ID     uint           `gorm:"primaryKey"`
	Dishes pq.StringArray `gorm:"type:text[]"`
}

func connectToRestaurantService() (gRPCRestaurant.ProcessCreaterClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Не удалось подключиться к RestaurantService: %v", err)
	}

	client := gRPCRestaurant.NewProcessCreaterClient(conn)
	return client, conn
}

type server struct {
	gRPCOrder.UnimplementedOrderCreaterServer
}

func (s *server) Create(ctx context.Context, req *gRPCOrder.OrderRequest) (*gRPCOrder.OrderResponse, error) {
	for _, elem := range req.Dishes {
		fmt.Println(elem)
	}

	orderDetails := &OrderDetails{
		Dishes: req.Dishes,
	}

	if err := db.Create(&orderDetails).Error; err != nil {
		log.Fatalf("Ошибка при сохранении заказа в базу данных: %v", err)
	}

	fmt.Printf("Присвоенный ID: %d\n", orderDetails.ID)

	restaurantClient, conn := connectToRestaurantService()
	defer conn.Close()

	restaurantReq := &gRPCRestaurant.OrderDetails{
		OrderID: int32(orderDetails.ID),
		Dishes:  req.Dishes,
	}

	ctxRestaurant, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	response, err := restaurantClient.Create(ctxRestaurant, restaurantReq)
	if err != nil {
		log.Fatalf("Ошибка при вызове ProcessOrder: %v", err)
	}

	log.Printf("Получен статус от RestaurantService: %s", response.Status)

	return &gRPCOrder.OrderResponse{OrderID: int32(orderDetails.ID)}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}

	if err := db.AutoMigrate(&OrderDetails{}); err != nil {
		log.Fatalf("Не удалось выполнить миграцию: %v", err)
	}

	grpcServer := grpc.NewServer()

	gRPCOrder.RegisterOrderCreaterServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	log.Println("gRPC сервер запущен на порту :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
