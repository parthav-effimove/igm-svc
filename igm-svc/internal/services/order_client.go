package services

import (
	"context"
	"fmt"
)

type OrderDetails struct{
	OrderID string
	UserID string
	BPPID string
	BPPURI string
	ProviderID string
	Status string
}

type OrderClient interface{
	VerifyOrder(ctx context.Context,orderID,userID string)(*OrderDetails,error)
}

type mockOrderClient struct{}

func NewMockOrderClient() OrderClient{
	return &mockOrderClient{}
}

func(c *mockOrderClient)VerifyOrder(ctx context.Context,orderID,userID string)(*OrderDetails,error){
	//TODO add grpc call to order service

	if orderID==""{
		return nil,fmt.Errorf("order_id is required")
	}
	return &OrderDetails{
		OrderID: orderID,
		UserID: userID,
		BPPID: "prepod.in",
		BPPURI: "http://prepod",
		ProviderID: "MP@",
		Status: "Order-delivered",
	},nil
}