package skywalking

import (
	"fmt"
	"strings"

	apmclient "github.com/CloudDetail/apo-module/apm/client/v1"
	"github.com/CloudDetail/apo-module/apm/model/v1"
	"github.com/CloudDetail/apo-module/apm/model/v1/transform"
)

var otSpanTagsMapping = map[string]string{
	"url":         model.AttributeURLFULL,
	"status_code": model.AttributeHTTPStatusCode,
}

var otSpanLogsMapping = map[string]string{
	"error.kind": model.AttributeExceptionType,
	"message":    model.AttributeExceptionMessage,
	"stack":      model.AttributeExceptionStacktrace,
}

func ConvertToServiceNodes(swTrace *SkywalkingTrace) ([]*model.OtelServiceNode, error) {
	traceData := model.NewOTelTrace("skywalking")
	if swTrace.Spans == nil && len(swTrace.Spans) == 0 {
		return traceData.GetServiceNodes(), nil
	}

	traceTree := model.NewOtelTree()
	for _, swSpan := range swTrace.Spans {
		if otelSpan := swSpanToSpan(swSpan); otelSpan != nil {
			if err := traceTree.AddSpan(otelSpan); err != nil {
				return nil, err
			}
		}
	}

	if err := traceTree.BuildRelation4Spans(traceData); err != nil {
		return nil, err
	}
	return traceData.GetServiceNodes(), nil
}

func swSpanToSpan(span *SkywalkingSpan) *model.OtelSpan {
	dest := model.NewOtelSpan()
	dest.SetSpanId(transform.SegmentIDToSpanID(span.SegmentId, uint32(span.SpanId)))
	dest.SetOriginalSpanId("SKYWALKING", fmt.Sprintf("%s-%d", span.SegmentId, span.SpanId))

	// parent spanid = -1, means(root span) no parent span in current skywalking segment, so it is necessary to search for the parent segment.
	if span.ParentSpanId != -1 {
		dest.SetParentSpanId(transform.SegmentIDToSpanID(span.SegmentId, uint32(span.ParentSpanId)))
	} else if len(span.Refs) == 1 {
		// TODO: SegmentReference references usually have only one element, but in batch consumer case, such as in MQ or async batch process, it could be multiple.
		// We only handle one element for now.
		dest.SetParentSpanId(transform.SegmentIDToSpanID(span.Refs[0].ParentSegmentId, uint32(span.Refs[0].ParentSpanId)))
	}

	dest.SetName(span.EndpointName)
	dest.SetServiceName(span.ServiceCode)
	dest.SetStartTime(span.StartTime * 1e6)                 // ms -> ns
	dest.SetDuration((span.EndTime - span.StartTime) * 1e6) // ms -> ns
	setInternalSpanStatus(span, dest)
	setInternalSpanKind(span, dest)
	if span.EndpointName == "UndertowDispatch" && span.SpanId == 0 {
		dest.SetKind(model.SpanKindServer)
	}

	if span.SpanType == SpanType_Local {
		if span.SpanLayer == SpanLayer_RPCFramework && span.Component == "GRPC" {
			return nil
		}
		if span.SpanLayer == SpanLayer_MQ && span.Component == "kafka-producer" {
			return nil
		}
	}
	swKvPairsToInternalAttributes(span, dest.Attributes) // Attributes
	swLogsToSpanEvents(span.Logs, dest)                  // Events
	return dest
}

func setInternalSpanStatus(span *SkywalkingSpan, dest *model.OtelSpan) {
	if span.IsError {
		dest.SetCode(model.StatusCodeError)
	} else {
		dest.SetCode(model.StatusCodeOk)
	}
}

func setInternalSpanKind(span *SkywalkingSpan, dest *model.OtelSpan) {
	switch {
	case span.SpanLayer == SpanLayer_MQ:
		if span.SpanType == SpanType_Entry {
			dest.SetKind(model.SpanKindConsumer)
		} else if span.SpanType == SpanType_Exit {
			dest.SetKind(model.SpanKindProducer)
		}
	case span.SpanType == SpanType_Exit:
		dest.SetKind(model.SpanKindClient)
	case span.SpanType == SpanType_Entry:
		dest.SetKind(model.SpanKindServer)
	case span.SpanType == SpanType_Local:
		dest.SetKind(model.SpanKindInternal)
	default:
		dest.SetKind(model.SpanKindUnspecified)
	}
}

func swLogsToSpanEvents(logs []*SkywalkingLogEntity, dest *model.OtelSpan) {
	if len(logs) == 0 {
		return
	}

	for _, log := range logs {
		if len(log.Data) > 0 {
			attributes := make(map[string]string)
			swKvPairsToInternalLogAttributes(log.Data, attributes)

			exceptionType, exist := attributes[model.AttributeExceptionType]
			if exist {
				message := attributes[model.AttributeExceptionMessage]
				stack := attributes[model.AttributeExceptionStacktrace]
				// ms -> us
				dest.AddException(log.Time*1000, exceptionType, message, stack)
			}
		}
	}
}

func swKvPairsToInternalAttributes(span *SkywalkingSpan, dest map[string]string) {
	if span.SpanType == SpanType_Exit {
		dest[model.AttributeNetPeerName] = span.Peer
		if span.SpanLayer == SpanLayer_RPCFramework {
			dest[model.AttributeRpcSystem] = strings.ToLower(getServer(span.Component))
		}
	}

	if len(span.Tags) == 0 {
		return
	}

	if span.SpanLayer == SpanLayer_MQ {
		if span.SpanType == SpanType_Exit || span.SpanType == SpanType_Entry {
			dest[model.AttributeMessageSystem] = strings.ToLower(getServer(span.Component))
		}
	}

	for _, pair := range span.Tags {
		if setCacheAttribute(dest, pair.Key, pair.Value, span.SpanType) &&
			setDbAttribute(dest, pair.Key, pair.Value) &&
			setMqAttribute(dest, pair.Key, pair.Value) {

			if otKey, ok := otSpanTagsMapping[pair.Key]; ok {
				dest[otKey] = pair.Value
			} else {
				dest[pair.Key] = pair.Value
			}
		}
	}
}

func setCacheAttribute(dest map[string]string, key string, value string, spanType SpanType) bool {
	if !strings.HasPrefix(key, "cache.") {
		return true
	}
	if spanType == SpanType_Local {
		// EhCache...
		dest[key] = value
	} else if key == "cache.type" {
		// Redis...
		// Xmemcached -> memcached
		dest[model.AttributeDBSystem] = strings.ToLower(getServer(value))
	} else if key == "cache.cmd" {
		// Redis、Memcached
		dest[model.AttributeDBStatement] = value
	} else if key == "cache.op" {
		// Aerospike
		dest[model.AttributeDBOperation] = value
	} else {
		dest[key] = value
	}
	return false
}

func setDbAttribute(dest map[string]string, key string, value string) bool {
	if !strings.HasPrefix(key, "db.") {
		return true
	}
	if key == "db.type" {
		// Mysql...
		dest[model.AttributeDBSystem] = strings.ToLower(getServer(value))
	} else if key == "db.instance" {
		dest[model.AttributeDBName] = value
	} else if key == "db.statement" {
		dest[model.AttributeDBStatement] = value
		if value != "" {
			operation, table := apmclient.SQLParseOperationAndTableNEW(value)
			if operation != "" {
				dest[model.AttributeDBOperation] = operation
				dest[model.AttributeDBSQLTable] = table
			}
		}
	} else {
		dest[key] = value
	}
	return false
}

func setMqAttribute(dest map[string]string, key string, value string) bool {
	if !strings.HasPrefix(key, "mq.") {
		return true
	}

	if key == "mq.queue" {
		dest[model.AttributeMessageDestinationName] = value
	} else if key == "mq.topic" {
		if _, exist := dest[model.AttributeMessageDestinationName]; !exist {
			dest[model.AttributeMessageDestinationName] = value
		}
	} else if key == "mq.broker" {
		dest[model.AttributeNetPeerName] = value
	} else {
		dest[key] = value
	}
	return false
}

func swKvPairsToInternalLogAttributes(pairs []*SkywalkingKeyValue, dest map[string]string) {
	if len(pairs) == 0 {
		return
	}

	for _, pair := range pairs {
		otKey, ok := otSpanLogsMapping[pair.Key]
		if ok {
			// Replace Tag to OTel Tag
			dest[otKey] = pair.Value
		} else {
			dest[pair.Key] = pair.Value
		}
	}
}
