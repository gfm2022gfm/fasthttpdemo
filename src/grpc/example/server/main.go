package main

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
)

const (
	// gRPC服务地址
	Address = "127.0.0.1:9988"
)

var sugarLogger *zap.SugaredLogger

type helloService struct{}

var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *HelloRequest) (*HelloResponse, error) {
	resp := new(HelloResponse)
	resp.Message = fmt.Sprintf("Hello %s.", in.Name)

	return resp, nil
}

func (h helloService) SayHi(ctx context.Context, in *HiRequest) (*HiResponse, error) {
	resp := new(HiResponse)
	resp.Message = fmt.Sprintf("Hi %s, grade=%d, school=%s, grade=%d, status=%d", in.Name, in.Grade, in.School, in.Grade, in.Status)
	sugarLogger.Info("grpc id : ", in.GetAge())
	return resp, nil
}

func main() {
	InitLogger()
	defer sugarLogger.Sync()
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	RegisterHelloServer(s, HelloService)
	fmt.Println("Listen on " + Address)
	grpclog.Println("Listen on " + Address)
	s.Serve(listen)
}

func InitLogger() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller())
	sugarLogger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./logs/rpc.log",
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}

	return zapcore.AddSync(lumberJackLogger)
}
