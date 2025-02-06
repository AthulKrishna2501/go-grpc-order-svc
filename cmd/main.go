package main

import (
	"fmt"
	"log"
	"net"

	"github.com/AthulKrishna2501/go-grpc-order-svc/pkg/clients"
	"github.com/AthulKrishna2501/go-grpc-order-svc/pkg/config"
	"github.com/AthulKrishna2501/go-grpc-order-svc/pkg/db"
	"github.com/AthulKrishna2501/go-grpc-order-svc/pkg/pb"
	"google.golang.org/grpc"
)

func main() {
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("Failed to load config:", err)
	}

	h := db.Init(c.DBUrl)

	lis, err := net.Listen("tcp", c.Port)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	fmt.Println("Order Service running on", c.Port)

	orderService := &clients.Server{
		H:              h,
		CartService:    clients.InitCartServiceClient(c.CartSvcUrl),
		AddressService: clients.InitAddressServiceClient(c.AddressSvcUrl),
		ProductService: clients.InitProductServiceClient(c.ProductSvcUrl),
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, orderService)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("Failed to serve:", err)
	}
}
