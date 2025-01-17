package jaeger

import (
	"strconv"
	"strings"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

var otSpanLogsMapping = map[string]string{
	"error.kind": model.AttributeExceptionType,
	"message":    model.AttributeExceptionMessage,
	"stack":      model.AttributeExceptionStacktrace,
}

var otSpanTagsIgnoreMapping = map[string]bool{
	"internal.span.format": true,
	"otel.library.name":    true,
	"otel.library.version": true,
	"otel.scope.name":      true,
	"otel.scope.version":   true,
	"thread.id":            true,
	"thread.name":          true,
}

func ConvertToServiceNodes(jaegerData *JaegerData) ([]*model.OtelServiceNode, error) {
	traceData := model.NewOTelTrace("otel")
	if jaegerData.Spans == nil && len(jaegerData.Spans) == 0 || len(jaegerData.Processes) == 0 {
		return traceData.GetServiceNodes(), nil
	}

	processServiceNameMap := make(map[string]string)
	for key, process := range jaegerData.Processes {
		processServiceNameMap[key] = process.ServiceName
	}

	traceTree := model.NewOtelTree()
	for _, span := range jaegerData.Spans {
		serviceName := processServiceNameMap[span.ProcessID]
		if err := traceTree.AddSpan(jSpanToInternal(span, serviceName)); err != nil {
			return nil, err
		}
	}
	if err := traceTree.BuildRelation4Spans(traceData); err != nil {
		return nil, err
	}
	return traceData.GetServiceNodes(), nil
}

func jSpanToInternal(span *JaegerSpan, serviceName string) *model.OtelSpan {
	dest := model.NewOtelSpan()
	dest.SetSpanId(span.SpanId)
	dest.SetOriginalSpanId("OTEL", span.SpanId)
	dest.SetServiceName(serviceName)
	dest.SetName(span.OperationName)
	dest.SetStartTime(span.StartTime * 1000) // us -> ns
	dest.SetDuration(span.Duration * 1000)   // us -> ns

	parentSpanID := span.GetParentSpanID()
	if len(parentSpanID) > 0 {
		dest.SetParentSpanId(parentSpanID)
	}
	jTagsToInternalAttributes(span.Tags, dest.Attributes)

	if spanKindAttr, ok := dest.Attributes[TagSpanKind]; ok {
		dest.SetKind(jSpanKindToInternal(spanKindAttr))
		delete(dest.Attributes, TagSpanKind)
	}
	if _, ok := dest.Attributes["sw8.segment_id"]; ok && span.OperationName == "UndertowDispatch" {
		// FIX Mismatch SpanId for UndertowDispatch.
		dest.SetKind(model.SpanKindServer)
	}
	setInternalSpanStatus(dest.Attributes, dest)

	jLogsToSpanExceptions(span.Logs, dest)

	return dest
}

func jTagsToInternalAttributes(tags []*JaegerKeyValue, dest map[string]string) {
	for _, tag := range tags {
		if _, exist := otSpanTagsIgnoreMapping[tag.Key]; !exist {
			dest[tag.Key] = getTagStrValue(tag)
		}
	}
}

func getTagStrValue(tag *JaegerKeyValue) string {
	switch tag.Type {
	case "string":
		return tag.GetVStr()
	case "bool":
		return strconv.FormatBool(tag.GetVBool())
	case "int64":
		return strconv.FormatInt(tag.GetVInt64(), 10)
	case "float64":
		return strconv.FormatFloat(tag.GetVFloat64(), 'b', 5, 32)
	default:
		// Ignore Other Types
		return ""
	}
}

func jTagsToInternalLogAttributes(tags []*JaegerKeyValue, dest map[string]string) {
	if len(tags) == 0 {
		return
	}

	for _, tag := range tags {
		otKey, ok := otSpanLogsMapping[tag.Key]
		if ok {
			// Replace Tag to OTel Tag
			dest[otKey] = getTagStrValue(tag)
		} else {
			dest[tag.Key] = getTagStrValue(tag)
		}
	}
}

func setInternalSpanStatus(attrs map[string]string, span *model.OtelSpan) {
	statusCode := model.StatusCodeUnset
	statusExists := false

	if errorVal, ok := attrs[TagError]; ok && errorVal == "true" {
		statusCode = model.StatusCodeError
		delete(attrs, TagError)
		statusExists = true
	}

	if codeAttr, ok := attrs[OtelStatusCode]; ok {
		if !statusExists {
			// The error tag is the ultimate truth for a Jaeger spans' error
			// status. Only parse the otel.status_code tag if the error tag is
			// not set to true.
			statusExists = true
			switch strings.ToUpper(codeAttr) {
			case "OK":
				statusCode = model.StatusCodeOk
			case "ERROR":
				statusCode = model.StatusCodeError
			}
		}
		delete(attrs, OtelStatusCode)
	} else if httpCodeAttr, ok := attrs[model.AttributeHTTPStatusCode]; !statusExists && ok {
		// Fallback to introspecting if this span represents a failed HTTP
		// request or response, but again, only do so if the `error` tag was
		// not set to true and no explicit status was sent.
		if code, err := getStatusCodeFromHTTPStatusAttr(httpCodeAttr, span.Kind); err == nil {
			if code != model.StatusCodeUnset {
				statusExists = true
				statusCode = code
			}
		}
	}

	if statusExists {
		span.SetCode(statusCode)
	}
}

func getStatusCodeFromHTTPStatusAttr(attrVal string, kind model.OtelSpanKind) (model.OtelStatusCode, error) {
	statusCode, err := strconv.ParseInt(attrVal, 10, 0)
	if err != nil {
		return model.StatusCodeUnset, err
	}

	// For HTTP status codes in the 4xx range span status MUST be left unset
	// in case of SpanKind.SERVER and MUST be set to Error in case of SpanKind.CLIENT.
	// For HTTP status codes in the 5xx range, as well as any other code the client
	// failed to interpret, span status MUST be set to Error.
	if statusCode >= 400 && statusCode < 500 {
		switch kind {
		case model.SpanKindClient:
			return model.StatusCodeError, nil
		case model.SpanKindServer:
			return model.StatusCodeUnset, nil
		}
	}

	return StatusCodeFromHTTP(statusCode), nil
}

func jSpanKindToInternal(spanKind string) model.OtelSpanKind {
	switch spanKind {
	case "client":
		return model.SpanKindClient
	case "server":
		return model.SpanKindServer
	case "producer":
		return model.SpanKindProducer
	case "consumer":
		return model.SpanKindConsumer
	case "internal":
		return model.SpanKindInternal
	}
	return model.SpanKindUnspecified
}

func jLogsToSpanExceptions(logs []*JaegerLog, dest *model.OtelSpan) {
	if len(logs) == 0 {
		return
	}

	for _, log := range logs {
		if len(log.Fields) > 0 {
			attributes := make(map[string]string)
			jTagsToInternalLogAttributes(log.Fields, attributes)

			exceptionType, exist := attributes[model.AttributeExceptionType]
			if exist {
				message := attributes[model.AttributeExceptionMessage]
				stack := attributes[model.AttributeExceptionStacktrace]
				// us
				dest.AddException(log.Timestamp, exceptionType, message, stack)
			}
		}
	}
}
