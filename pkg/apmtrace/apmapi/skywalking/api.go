package skywalking

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unsafe"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

type SkywalkingApi struct {
	Address string
	Token   string
	Timeout time.Duration
}

func getToken(user string, password string) string {
	base := user + ":" + password
	baseStr := unsafe.Slice(unsafe.StringData(base), len(base))
	return "Basic " + base64.StdEncoding.EncodeToString(baseStr)
}

func NewSkywalkingApi(address string, user string, passwd string, timeout int64) *SkywalkingApi {
	token := ""
	if len(user)+len(passwd) > 0 {
		token = getToken(user, passwd)
	}
	return &SkywalkingApi{
		Address: fmt.Sprintf("http://%s/graphql", address),
		Timeout: time.Duration(timeout) * time.Second,
		Token:   token,
	}
}

func (sw *SkywalkingApi) QueryList(traceId string, startTimeMs int64, attributes string) ([]*model.OtelServiceNode, error) {
	requestBody := fmt.Sprintf(`{"query": "query queryTrace($traceId: ID!) {trace: queryTrace(traceId: $traceId) {spans{traceId segmentId spanId parentSpanId refs{traceId parentSegmentId parentSpanId type} serviceCode serviceInstanceName startTime endTime endpointName type peer component isError layer tags{key value} logs{time data {key value}}}}}","variables": {"traceId": "%s"}}`, traceId)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if len(sw.Token) > 0 {
		headers["Authorization"] = sw.Token
	}
	resp, err := queryJson(sw.Address, headers, requestBody, sw.Timeout)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("[x Not Authorized] Please specify username and password")
	}
	defer resp.Body.Close()
	var response SkywalkingResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if len(response.Data.Trace.Spans) == 0 {
		return nil, fmt.Errorf("[x Trace NotFound] Skywalking traceId: %s", traceId)
	}

	return ConvertToServiceNodes(&response.Data.Trace)
}

func queryJson(requestUrl string, headers map[string]string, body string, timeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, requestUrl, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: timeout,
	}
	return client.Do(req)
}
