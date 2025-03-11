package httpserver

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/CloudDetail/apo-apm-adapter/pkg/global"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/pprof"
)

func StartHttpServer(port int) {
	app := iris.Default()

	app.Post("/trace/list", queryTraceList)

	p := pprof.New()
	app.Any("/debug/pprof", p)
	app.Any("/debug/pprof/{action:path}", p)

	// Graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		log.Println("Shutting down HTTP server...")
		_ = app.Shutdown(context.Background())
	}()

	err := app.Listen(":" + strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Failed to start the http server %v", err)
	}
}

type BasicStatus string

const (
	Success BasicStatus = "success"
	Failure BasicStatus = "failure"
)

type BasicResponse struct {
	Status  BasicStatus `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func queryTraceList(ctx iris.Context) {
	var request TraceListRequest
	if err := ctx.ReadJSON(&request); err != nil {
		responseWithError(ctx, err)
		return
	}

	result, err := global.TRACE_CLIENT.QueryTraceList(request.ApmType, request.TraceId, request.StartTime, request.Attributes)
	if err != nil {
		log.Printf("[QueryTraceList] apmType: %s, traceId: %s, error: %v", request.ApmType, request.TraceId, err)
		responseWithError(ctx, err)
		return
	}
	log.Printf("[QueryTraceList] apmType: %s, traceId: %s, size: %d", request.ApmType, request.TraceId, len(result))
	ctx.JSON(iris.Map{
		"success": true,
		"data":    result,
	})
}

func responseWithError(ctx iris.Context, err error) {
	ctx.StopWithStatus(iris.StatusInternalServerError)
	ctx.JSON(iris.Map{
		"success":  false,
		"errorMsg": err.Error(),
	})
}

type TraceListRequest struct {
	ApmType    string `json:"apmType"`
	TraceId    string `json:"traceId"`
	StartTime  int64  `json:"startTime"`
	Attributes string `json:"attributes"`
}
