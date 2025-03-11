package elastic

import (
	"net/http"
	"strings"
	"time"

	"github.com/CloudDetail/apo-module/apm/model/v1"
	"github.com/elastic/go-elasticsearch/v7"
)

type ELASTICApi struct {
	ESClient
}

func NewELASTICApi(addr string, username string, password string, timeout int64) (esapi *ELASTICApi, err error) {
	var esaddr string
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		esaddr = addr
	} else {
		esaddr = "http://" + addr
	}

	cfg := elasticsearch.Config{
		Addresses: []string{esaddr},
		Username:  username,
		Password:  password,
	}

	transport := &http.Transport{}
	if timeout > 0 {
		transport.ResponseHeaderTimeout = time.Duration(timeout) * time.Second
	}

	cfg.Transport = transport

	esapi = &ELASTICApi{}
	esapi.es, err = elasticsearch.NewClient(cfg)

	return esapi, err
}

func (api *ELASTICApi) QueryList(traceId string, startTimeMs int64, attributes string) ([]*model.OtelServiceNode, error) {
	searchResp, err := api.searchSpans(traceId, "apm-*-span", "apm-*-transaction", "apm-*-error")
	if err != nil {
		return nil, err
	}

	return ConvertToServiceNodes(searchResp)
}
