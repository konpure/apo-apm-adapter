package jaeger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

type JaegerApi struct {
	Address string
	Timeout time.Duration
}

func NewJaegerApi(address string, timeout int64) *JaegerApi {
	return &JaegerApi{
		Address: fmt.Sprintf("http://%s/api/traces", address),
		Timeout: time.Duration(timeout) * time.Second,
	}
}

func (jaeger *JaegerApi) QueryList(traceId string, startTimeMs int64, attributes string) ([]*model.OtelServiceNode, error) {
	resp, err := queryJson(fmt.Sprintf("%s/%s", jaeger.Address, traceId), jaeger.Timeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response JaegerResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Data) == 0 {
		return nil, fmt.Errorf("[x Trace NotFound] Jaeger traceId: %s", traceId)
	}
	return ConvertToServiceNodes(&response.Data[0])
}

func queryJson(url string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	return client.Get(url)
}
