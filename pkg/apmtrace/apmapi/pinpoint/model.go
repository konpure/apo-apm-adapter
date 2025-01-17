package pinpoint

import (
	"strings"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

var otSpanTagsMapping = map[string]string{
	"Servlet Process":  model.AttributeHTTPURL,
	"http.status.code": model.AttributeHTTPStatusCode,
}

type PinpointResponse struct {
	TraceId    string             `json:"transactionId"`
	Complete   string             `json:"completeState"`
	CallStacks [][]interface{}    `json:"callStack"`
	Exception  *PinpointException `json:"exception"`
}

type PinpointException struct {
	Stacktrace string `json:"stacktrace"`
	Message    string `json:"message"`
}

func (resp *PinpointResponse) ConvertToServiceNodes() ([]*model.OtelServiceNode, error) {
	traceTree := model.NewOtelTree()
	spanMap := make(map[string]*model.OtelSpan, 0)
	childrenSpans := make(map[string][]*model.OtelSpan, 0)
	clientServerSpans := make(map[string]bool, 0)
	var rootSpan *model.OtelSpan

	for _, callStack := range resp.CallStacks {
		parentSpanId := callStack[7].(string)
		if callStack[22].(bool) {
			// HasException
			parentSpan := spanMap[parentSpanId]
			parentSpan.SetCode(model.StatusCodeError)
			parentSpan.AddException(parentSpan.StartTime/1000, callStack[10].(string), callStack[11].(string), "")
			markEntrySpanError(spanMap, parentSpan)
			continue
		}

		if !callStack[8].(bool) {
			// Attributes
			attributeKey := callStack[10].(string)
			if otKey, ok := otSpanTagsMapping[attributeKey]; ok {
				spanMap[parentSpanId].AddAttribute(otKey, callStack[11].(string))
			}
			continue
		}
		span := model.NewOtelSpan()
		span.SetStartTime(uint64(callStack[1].(float64)) * 1000000)
		span.SetDuration(uint64(callStack[2].(float64)-callStack[1].(float64)) * 1000000)
		span.SetServiceName(callStack[4].(string))
		span.SetName(getSpanName(callStack[17].(string), callStack[10].(string)))
		span.SetSpanId(callStack[6].(string))
		span.SetParentSpanId(parentSpanId)
		if callStack[9].(bool) {
			// IsServer
			span.SetKind(model.SpanKindServer)
			url := callStack[11].(string)
			if url != "" {
				span.AddAttribute(model.AttributeHTTPURL, callStack[11].(string))
				span.SetName(url)
			}
			if span.PSpanId != "" {
				spanMap[span.PSpanId].SetKind(model.SpanKindClient)
				clientServerSpans[span.PSpanId] = true
			}
		} else {
			span.SetKind(getSpanKind(callStack[19].(string)))
			if span.Kind.IsExit() {
				title := callStack[11].(string)
				if strings.Contains(title, "://") {
					// http、dubbo、mysql
					span.AddAttribute(model.AttributeHTTPURL, title)
				}
			}
		}
		spanMap[span.SpanId] = span

		if span.Kind.IsEntry() && span.PSpanId == "" {
			rootSpan = span
		} else {
			children, exist := childrenSpans[span.PSpanId]
			if !exist {
				children = make([]*model.OtelSpan, 0)
			}
			children = append(children, span)
			childrenSpans[span.PSpanId] = children
		}
	}

	checkClientServerSpans(spanMap, childrenSpans, clientServerSpans)
	checkMiddlewareSpans(childrenSpans, clientServerSpans, rootSpan)

	addSpanToTree(traceTree, childrenSpans, rootSpan)
	traceData := model.NewOTelTrace("pinpoint")
	if err := traceTree.BuildRelation4Spans(traceData); err != nil {
		return nil, err
	}
	return traceData.GetServiceNodes(), nil
}

func checkClientServerSpans(
	spanMap map[string]*model.OtelSpan,
	childrenSpans map[string][]*model.OtelSpan,
	clientServerSpans map[string]bool) {

	for spanId := range clientServerSpans {
		setParentExitSpanInternal(spanMap, spanMap[spanId].PSpanId)
	}

	for spanId := range clientServerSpans {
		clientServerSpan := spanMap[spanId]
		childrens := childrenSpans[spanId]
		if len(childrens) > 1 {
			for _, child := range childrens {
				if child.Kind.IsEntry() {
					childrenSpans[spanId] = []*model.OtelSpan{child}
				} else {
					collectChildInfo(childrenSpans, child, clientServerSpan)
				}
			}
		}
	}
}

func checkMiddlewareSpans(childrenSpans map[string][]*model.OtelSpan, clientServerSpans map[string]bool, parentSpan *model.OtelSpan) {
	childrens := childrenSpans[parentSpan.SpanId]
	if parentSpan.Kind.IsExit() && !clientServerSpans[parentSpan.SpanId] {
		childrenSpans[parentSpan.SpanId] = []*model.OtelSpan{}
		for _, child := range childrens {
			collectChildInfo(childrenSpans, child, parentSpan)
		}
		return
	}

	for _, child := range childrens {
		checkMiddlewareSpans(childrenSpans, clientServerSpans, child)
	}
}

func addSpanToTree(traceTree *model.OtelTree, childrenSpans map[string][]*model.OtelSpan, parentSpan *model.OtelSpan) {
	traceTree.AddSpan(parentSpan)

	for _, childSpan := range childrenSpans[parentSpan.SpanId] {
		addSpanToTree(traceTree, childrenSpans, childSpan)
	}
}

func setParentExitSpanInternal(spanMap map[string]*model.OtelSpan, parentSpanId string) {
	parentSpan := spanMap[parentSpanId]
	if parentSpan == nil || parentSpan.Kind.IsEntry() {
		return
	}
	if parentSpan.Kind.IsExit() {
		parentSpan.SetKind(model.SpanKindInternal)
	}
	setParentExitSpanInternal(spanMap, parentSpan.PSpanId)
}

func collectChildInfo(childrenSpans map[string][]*model.OtelSpan, parentSpan *model.OtelSpan, clientSpan *model.OtelSpan) {
	if len(parentSpan.Exceptions) > 0 {
		for _, exception := range parentSpan.Exceptions {
			clientSpan.SetCode(model.StatusCodeError)
			clientSpan.AddException(exception.Timestamp, exception.Type, exception.Message, exception.Stack)
		}
	}
	if len(parentSpan.Attributes) > 0 {
		for key, value := range parentSpan.Attributes {
			clientSpan.AddAttribute(key, value)
		}
	}
	for _, child := range childrenSpans[parentSpan.SpanId] {
		collectChildInfo(childrenSpans, child, clientSpan)
	}
}

func markEntrySpanError(spanMap map[string]*model.OtelSpan, parentSpan *model.OtelSpan) {
	if parentSpan == nil {
		return
	}
	if parentSpan.Kind.IsEntry() {
		parentSpan.SetCode(model.StatusCodeError)
		return
	}
	markEntrySpanError(spanMap, spanMap[parentSpan.PSpanId])
}

func getSpanName(className string, methodName string) string {
	if className == "" {
		return methodName
	}
	return className + "." + methodName
}

func getSpanKind(serviceType string) model.OtelSpanKind {
	if IsClient(serviceType) {
		return model.SpanKindClient
	}
	return model.SpanKindInternal
}
