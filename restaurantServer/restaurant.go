package main

import (
	"context"
	"fmt"
	gRPCRestaurant "gRPCProcessingServer/restaurantService"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var orderChannel chan *OrderStatus

type server struct {
	gRPCRestaurant.UnimplementedProcessCreaterServer
}

type OrderStatus struct {
	ID     uint `gorm:"primaryKey"`
	Status string
}

func worker(id int, orders <-chan *OrderStatus) {
	for order := range orders {
		log.Printf("Worker %d начал обработку заказа %d", id, order.ID)
		order.work()
		log.Printf("Worker %d завершил обработку заказа %d", id, order.ID)
	}
}

func (order *OrderStatus) work() {
	time.Sleep(time.Duration(rand.IntN(3)+5) * time.Second)
	order.Status = "Готовится"
	if err := db.Save(order).Error; err != nil {
		log.Printf("Ошибка при обновлении статуса заказа: %v", err)
	}
	log.Printf("Статус заказа %d изменен на 'Готовится'", order.ID)

	time.Sleep(time.Duration(rand.IntN(7)+7) * time.Second)
	order.Status = "Готов к выдаче"
	if err := db.Save(order).Error; err != nil {
		log.Printf("Ошибка при обновлении статуса заказа: %v", err)
	}
	log.Printf("Статус заказа %d изменен на 'Завершен'", order.ID)

}

func (s *server) Create(ctx context.Context, req *gRPCRestaurant.OrderDetails) (*gRPCRestaurant.OrderStatus, error) {
	fmt.Printf("Заказ с id %d получен\n", req.OrderID)

	orderStatus := OrderStatus{Status: "Принят в работу"}

	if err := db.Create(&orderStatus).Error; err != nil {
		log.Fatalf("Ошибка при сохранении заказа в базу данных: %v", err)
	}

	restaurantResponse := &gRPCRestaurant.OrderStatus{
		OrderID: int32(orderStatus.ID),
		Status:  orderStatus.Status,
	}
	go func() {
		orderChannel <- &orderStatus
	}()

	return restaurantResponse, nil
}

func (s *server) Status(ctx context.Context, req *gRPCRestaurant.OrderID) (*gRPCRestaurant.OrderStatus, error) {
	var orderStatus OrderStatus
	if err := db.First(&orderStatus, req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении статуса заказа: %v", err)
	}

	restaurantResponse := &gRPCRestaurant.OrderStatus{
		OrderID: int32(orderStatus.ID),
		Status:  orderStatus.Status,
	}

	return restaurantResponse, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
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

	if err := db.AutoMigrate(&OrderStatus{}); err != nil {
		log.Fatalf("Не удалось выполнить миграцию: %v", err)
	}

	orderChannel = make(chan *OrderStatus, 100)

	for i := 1; i <= 3; i++ {
		go worker(i, orderChannel)
	}

	grpcServer := grpc.NewServer()

	gRPCRestaurant.RegisterProcessCreaterServer(grpcServer, &server{})
	reflection.Register(grpcServer)
	log.Println("gRPC сервер запущен на порту :50050")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
