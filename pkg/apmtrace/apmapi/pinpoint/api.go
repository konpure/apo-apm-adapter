package pinpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

type PinpointApi struct {
	Address string
	Timeout time.Duration
}

func NewPinpointApi(address string, timeout int64) (ppApi *PinpointApi, err error) {
	return &PinpointApi{
		Address: fmt.Sprintf("http://%s/transactionInfo.pinpoint", address),
		Timeout: time.Duration(timeout) * time.Second,
	}, nil
}

func (pinpoint *PinpointApi) QueryList(traceId string, startTimeMs int64, attributes string) ([]*model.OtelServiceNode, error) {
	client := &http.Client{
		Timeout: pinpoint.Timeout,
	}
	resp, err := client.Get(fmt.Sprintf("%s?traceId=%s", pinpoint.Address, strings.ReplaceAll(traceId, "^", "%5E")))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response PinpointResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if response.Exception != nil {
		return nil, fmt.Errorf("[x Trace NotFound] Pinpoint traceId: %s", traceId)
	}
	if response.Complete != "Complete" {
		return nil, fmt.Errorf("[x Trace NotComplete] Pinpoint traceId: %s", traceId)
	}
	return response.ConvertToServiceNodes()
}
