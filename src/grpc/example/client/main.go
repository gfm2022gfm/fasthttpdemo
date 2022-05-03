package main

import (
	"example"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	// gRPC服务地址
	Address = "127.0.0.1:9988"
)

func main() {
	conn, err := grpc.Dial(Address, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalln(err)
	}
	defer conn.Close()

	c := example.NewHelloClient(conn)

	req := &example.HelloRequest{Name: "grpc"}
	res, err := c.SayHello(context.Background(), req)
	if err != nil {
		grpclog.Fatalln(err)
	}

	fmt.Println(res.Message)

	req2 := &example.HiRequest{Name: "grpc", Grade: 3, Age: 10, Status: 2, School: "zhuhai"}
	res2, err := c.SayHi(context.Background(), req2)
	if err != nil {
		grpclog.Fatalln(err)
	}

	fmt.Println(res2.Message)
}
