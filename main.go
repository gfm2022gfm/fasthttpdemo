package main

import (
	"flag"
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/natefinch/lumberjack"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	addr     = flag.String("addr", ":8080", "TCP address to listen to")
	compress = flag.Bool("compress", false, "Whether to enable transparent response compression")
)
var sugarLogger *zap.SugaredLogger

func main() {
	InitLogger()
	defer sugarLogger.Sync()
	flag.Parse()
	router := fasthttprouter.New()
	router.GET("/", requestHandler)
	router.GET("/index2", Index)
	router.GET("/hello/:name", Hello)
	router.GET("/forward2backednok", forward2backednok) //delete
	router.GET("/f2b", f2b)

	router.GET("/f3b/:id", f3b)
	router.GET("/add/:id", add)

	//client
	//doRequest()

	if err := fasthttp.ListenAndServe(*addr, router.Handler); err != nil {
		sugarLogger.Fatalf("Error in ListenAndServe: %v", err)
	}
}

func doRequest() string {
	url := fmt.Sprintf("%v:8091", os.Getenv("forwardbackend"))
	//url := "http://www.google.com"
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)   // <- do not forget to release
	defer fasthttp.ReleaseResponse(resp) // <- do not forget to release

	req.SetRequestURI(url)

	fasthttp.Do(req, resp)

	bodyBytes := resp.Body()
	sugarLogger.Info("doRequest body: %V", zap.String("env:%v", string(bodyBytes)))
	// User-Agent: fasthttp
	// Body:
	return string(bodyBytes)
}

func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}

func add(ctx *fasthttp.RequestCtx) {
	sugarLogger.Info("id= ", ctx.UserValue("id"))
	fmt.Fprintf(ctx, "id= %s\n", ctx.UserValue("id"))
}

func forward2backednok(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, doRequest())
}

func f3b(ctx *fasthttp.RequestCtx) {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: PrintLocalDial,
		},
	}

	url := fmt.Sprintf("%v:8091/add/:%v", os.Getenv("forwardbackend"), ctx.UserValue("id")) // os.Getenv("forwardbackend") + "/" + string(ctx.UserValue("name"))

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
		sugarLogger.Error("f3b 请求失败: %V", zap.String("%v", err.Error()))
		return
	}
	buf, err := ioutil.ReadAll(resp.Body)

	sugarLogger.Info("f3b 请求ok: %V", zap.String("%v", string(buf)))
	if err := resp.Body.Close(); err != nil {
		sugarLogger.Error("f3b 请求失败: %V", zap.String("%v", err.Error()))
	}
}

var fastHttpClient = map[string]*fasthttp.Client{}
var lock sync.RWMutex

func f2b(ctx *fasthttp.RequestCtx) {

	url := fmt.Sprintf("%v:8091", os.Getenv("forwardbackend"))
	sugarLogger.Info("f2b env: %V", zap.String("env:%v", url))

	status, resp, err := fasthttp.Get(nil, url)
	if err != nil {
		sugarLogger.Error("f2b 请求失败: %V", zap.String("%v", err.Error()))
		return
	}

	if status != fasthttp.StatusOK {
		sugarLogger.Error("f2b 请求没有成功: %V", zap.String("%v", string(status)))
		return
	}

	sugarLogger.Info("f2b body: %V", zap.String("env:%v", string(resp)))

	fmt.Fprintf(ctx, string(resp))
}

func PrintLocalDial(network, addr string) (net.Conn, error) {
	dial := net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	conn, err := dial.Dial(network, addr)
	if err != nil {
		return conn, err
	}
	sugarLogger.Info("connect done, use", conn.LocalAddr().String())
	return conn, err
}

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome!\n")
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	sugarLogger.Info("---------------- HTTP header 每一个键值对-------------")
	fmt.Fprintf(ctx, "Hello, world!\n v2.3\n")
	fmt.Fprintf(ctx, "env : %v \n\n", os.Getenv("forwardbackend"))

	fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
	fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
	fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
	fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
	fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
	fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
	fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
	fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
	fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
	fmt.Fprintf(ctx, "Your ip is %q\n\n", ctx.RemoteIP())

	fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)

	ctx.SetContentType("text/plain; charset=utf8")

	// Set arbitrary headers
	ctx.Response.Header.Set("X-My-Header", "my-header-value")

	// Set cookies
	var c fasthttp.Cookie
	c.SetKey("cookie-name")
	c.SetValue("cookie-value")
	ctx.Response.Header.SetCookie(&c)
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
		Filename:   "./logs/test.log",
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}

	return zapcore.AddSync(lumberJackLogger)
}
