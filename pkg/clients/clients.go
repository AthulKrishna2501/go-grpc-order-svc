package clients

import (
	"context"
	"errors"

	addrPb "github.com/AthulKrishna2501/go-grpc-address-svc/pkg/pb"
	cartPb "github.com/AthulKrishna2501/go-grpc-cart-service/pkg/pb"
	"github.com/AthulKrishna2501/go-grpc-order-svc/pkg/db"
	"github.com/AthulKrishna2501/go-grpc-order-svc/pkg/models"
	"github.com/AthulKrishna2501/go-grpc-order-svc/pkg/pb"
	proPb "github.com/AthulKrishna2501/go-grpc-product-svc/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	H              db.Handler
	CartService    cartPb.CartServiceClient
	AddressService addrPb.AddressServiceClient
	ProductService proPb.ProductServiceClient
}

func InitCartServiceClient(url string) cartPb.CartServiceClient {
	conn, _ := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return cartPb.NewCartServiceClient(conn)
}

func InitAddressServiceClient(url string) addrPb.AddressServiceClient {
	conn, _ := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return addrPb.NewAddressServiceClient(conn)
}

func InitProductServiceClient(url string) proPb.ProductServiceClient {
	conn, _ := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return proPb.NewProductServiceClient(conn)
}

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	cartRes, err := s.CartService.GetCart(ctx, &cartPb.GetCartRequest{UserId: req.UserId})
	if err != nil {
		return nil, err
	}

	addressRes, err := s.AddressService.GetAddress(ctx, &addrPb.GetAddressRequest{Id: req.UserId})
	if err != nil {
		return nil, err
	}

	var orderItems []models.OrderItem
	for _, cartItem := range cartRes.Items {
		orderItems = append(orderItems, models.OrderItem{
			ProductID:   cartItem.ProductId,
			ProductName: cartItem.ProductName,
			Price:       float64(cartItem.Price),
			Quantity:    int(cartItem.Quantity),
			TotalPrice:  cartItem.Price,
		})
	}

	order := models.Order{
		UserID:  req.UserId,
		Items:   orderItems,
		Address: addressRes.Address.District,
		Total:   calculateTotal(orderItems),
		Payment: "COD",
		Status:  "Pending",
	}

	for _, item := range cartRes.Items {
		_, err := s.ProductService.DecreaseStock(ctx, &proPb.DecreaseStockRequest{Id: item.ProductId})
		if err != nil {
			return nil, err
		}
	}

	if err := s.H.DB.Create(&order).Error; err != nil {
		return nil, errors.New("failed to create order")
	}

	return &pb.CreateOrderResponse{OrderId: int64(order.ID), Message: "Order Created", Success: true}, nil
}

func calculateTotal(items []models.OrderItem) float64 {
	var totalAmount float64
	for _, item := range items {
		totalAmount += item.Price
	}

	return totalAmount
}
