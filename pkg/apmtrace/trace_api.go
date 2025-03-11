package apmtrace

import (
	"errors"
	"fmt"
	"log"

	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi"
	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/elastic"
	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/jaeger"
	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/pinpoint"
	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/skywalking"
	"github.com/CloudDetail/apo-apm-adapter/pkg/config"
	"github.com/CloudDetail/apo-module/apm/model/v1"
)

const (
	APMTYPE_SW         = "skywalking"
	OTEL_EXPORT_JAEGER = "jaeger"
	APMTYPE_OTEL       = "otel"
	APMTYPE_ELASTIC    = "elastic"
	APMTYPE_PINPOINT   = "pinpoint"

	INVALID_API = "[x Build %s] %s is not set"
	VALID_API   = "[Build TraceApi] %s"
)

var (
	ErrNoAvaiableApmType error = errors.New("no match apmType is found")
)

type ApmTraceClient struct {
	apiMap map[string]apmapi.QueryByApmApi
}

func NewApmTraceClient(conf *config.TraceApiConfig, timeout int64) (*ApmTraceClient, error) {
	apiMap := make(map[string]apmapi.QueryByApmApi, 0)
	for _, apmType := range conf.ApmList {
		switch apmType {
		case APMTYPE_SW:
			buildSkywalkingApi(conf.Skywalking, apiMap, timeout)
		case OTEL_EXPORT_JAEGER:
			buildJaegerApi(conf.Jaeger, apiMap, timeout)
		case APMTYPE_ELASTIC:
			buildEsapmApi(conf.Elastic, apiMap, timeout)
		case APMTYPE_PINPOINT:
			buildPinpointApi(conf.Pinpoint, apiMap, timeout)
		default:
			log.Printf("Unknonw apmType: %s", apmType)
		}
	}

	if len(apiMap) == 0 {
		return nil, ErrNoAvaiableApmType
	}

	return &ApmTraceClient{
		apiMap: apiMap,
	}, nil
}

func buildSkywalkingApi(conf *config.SkywalkingConfig, apiMap map[string]apmapi.QueryByApmApi, timeout int64) {
	if conf == nil {
		log.Printf(INVALID_API, "SkywalkingApi", "skywalking")
		return
	}
	if len(conf.Address) == 0 {
		log.Printf(INVALID_API, "SkywalkingApi", "skywalking.address")
		return
	}
	log.Printf(VALID_API, "skywalking")
	apiMap[APMTYPE_SW] = skywalking.NewSkywalkingApi(conf.Address, conf.User, conf.Password, timeout)
}

func buildJaegerApi(conf *config.JaegerConfig, apiMap map[string]apmapi.QueryByApmApi, timeout int64) {
	if conf == nil {
		log.Printf(INVALID_API, "JaegerApi", "jaeger")
		return
	}
	if len(conf.Address) == 0 {
		log.Printf(INVALID_API, "JaegerApi", "jaeger.address")
		return
	}
	log.Printf(VALID_API, "jaeger")
	apiMap[APMTYPE_OTEL] = jaeger.NewJaegerApi(conf.Address, timeout)
}

func buildEsapmApi(conf *config.ElasticConfig, apiMap map[string]apmapi.QueryByApmApi, timeout int64) {
	if conf == nil {
		log.Printf(INVALID_API, "elasticApi", "elastic")
		return
	}
	if len(conf.Address) == 0 {
		log.Printf(INVALID_API, "elasticApi", "elastic")
		return
	}

	esAPMClient, err := elastic.NewELASTICApi(conf.Address, conf.User, conf.Password, timeout)
	if err != nil {
		log.Printf("[x Build elasticApi] %v", err)
		return
	}
	log.Printf(VALID_API, "elastic")
	apiMap[APMTYPE_ELASTIC] = esAPMClient
}

func buildPinpointApi(conf *config.PinpointConfig, apiMap map[string]apmapi.QueryByApmApi, timeout int64) {
	if conf == nil {
		log.Printf(INVALID_API, "pinpointApi", "pinpoint")
		return
	}
	if len(conf.Address) == 0 {
		log.Printf(INVALID_API, "pinpointApi", "pinpoint")
		return
	}

	ppAPMClient, err := pinpoint.NewPinpointApi(conf.Address, timeout)
	if err != nil {
		log.Printf("[x Build pinpointApi] %v", err)
		return
	}
	log.Printf(VALID_API, "pinpoint")
	apiMap[APMTYPE_PINPOINT] = ppAPMClient
}

func (client *ApmTraceClient) QueryTraceList(apmType string, traceId string, startTimeMs int64, attributes string) ([]*model.OtelServiceNode, error) {
	if api, exist := client.apiMap[apmType]; exist {
		return api.QueryList(traceId, startTimeMs, attributes)
	}
	return nil, fmt.Errorf("unknown apmType: %s", apmType)
}
