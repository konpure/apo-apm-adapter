package apmapi

import (
	"github.com/CloudDetail/apo-module/apm/model/v1"
)

type QueryByApmApi interface {
	QueryList(traceId string, startTimeMs int64) ([]*model.OtelServiceNode, error)
}
