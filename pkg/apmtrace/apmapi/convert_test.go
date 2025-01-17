package apmapi

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/elastic"
	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/jaeger"
	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/pinpoint"
	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace/apmapi/skywalking"
	"github.com/CloudDetail/apo-module/apm/model/v1"
)

func TestJaegerV1ConvertToTraceCases(t *testing.T) {
	testConvertToTraceCases(t,
		"jaeger-1.32",
		"http",
		"mysql",
		"redis",
		"dubbo",
		"grpc",
		"activemq",
		"rabbitmq",
		"rocketmq",
		"kafka",
	)
}

func TestJaegerConvertToTraceCases(t *testing.T) {
	testConvertToTraceCases(t,
		"jaeger",
		"error",
		"grpc",
		"http",
		"sw-otel",
		"mysql",
		"redis",
		"dubbo",
		"activemq",
		"rabbitmq",
		"rocketmq",
		"kafka",
	)
}

func TestSkywalkingConvertToTraceCases(t *testing.T) {
	testConvertToTraceCases(t,
		"skywalking",
		"error",
		"http",
		"redis",
		"dubbo",
		"grpc",
		"activemq",
		"rabbitmq",
		"rocketmq",
		"kafka",
	)
}

func TestESApmConvertToTraceCases(t *testing.T) {
	testConvertToTraceCases(t,
		"elastic",
		"http",
		"error",
		"dubbo",
	)
}

func TestPinPointConvertToTraceCases(t *testing.T) {
	testConvertToTraceCases(t,
		"pinpoint",
		"http",
		"db",
		"dubbo",
		"error",
	)
}

func testConvertToTraceCases(t *testing.T, apmType string, testCases ...string) {
	for _, testCase := range testCases {
		testTraceCase := buildTestTraceCase(t, apmType, testCase)
		if testTraceCase != nil {
			testConvertToTraceCase(t, apmType, testCase, testTraceCase)
		}
	}
}

func buildTestTraceCase(t *testing.T, apmType string, testCase string) *TestTraceCase {
	var (
		traceCase *TestTraceCase
		err       error
	)

	switch apmType {
	case "skywalking":
		{
			dataFile := fmt.Sprintf("testdata/tracelist/%s/%s/data.json", apmType, testCase)
			traceCase, err = convertSkywalkingToTraceCase(dataFile)
		}
	case "jaeger":
		{
			dataFile := fmt.Sprintf("testdata/tracelist/%s/%s/data.json", apmType, testCase)
			traceCase, err = convertJaegerToTraceCase(dataFile)
		}
	case "jaeger-1.32":
		{
			dataFile := fmt.Sprintf("testdata/tracelist/%s/%s/data.json", apmType, testCase)
			traceCase, err = convertJaegerToTraceCase(dataFile)
		}
	case "elastic":
		{
			dataFile := fmt.Sprintf("testdata/tracelist/%s/%s/data.json", apmType, testCase)
			traceCase, err = convertESApmToTraceCase(dataFile)
		}
	case "pinpoint":
		{
			dataFile := fmt.Sprintf("testdata/tracelist/%s/%s/data.json", apmType, testCase)
			traceCase, err = convertPinpointToTraceCase(dataFile)
		}
	default:
		err = fmt.Errorf("Unknown apmType: %s", apmType)
	}

	if err != nil {
		t.Errorf("Fail to convert to otel trace, Error: %v", err)
		return nil
	}

	return traceCase
}

func testConvertToTraceCase(t *testing.T, apmType string, testCase string, testTraceCase *TestTraceCase) {
	fileName := fmt.Sprintf("testdata/tracelist/%s/%s/validate.json", apmType, testCase)
	if isFileEmpty(fileName) {
		testTraceCase.Name = fmt.Sprintf("%s-%s", apmType, testCase)
		result, _ := json.Marshal(testTraceCase)
		fmt.Printf("%s", string(result))
		return
	}

	data, _ := os.ReadFile(fileName)
	traceCase := &TestTraceCase{}
	if err := json.Unmarshal(data, traceCase); err != nil {
		t.Errorf("Read json Failed, Error%v\n", err)
		return
	}

	checkIntEqual(t, "ServiceNode Size", len(traceCase.Services), len(testTraceCase.Services))
	for i, gotService := range testTraceCase.Services {
		expectService := traceCase.Services[i]
		checkServiceNode(t, expectService, gotService)
	}
}

func checkServiceNode(t *testing.T, expect *model.OtelServiceNode, got *model.OtelServiceNode) {
	checkIntEqual(t, "EntrySpans Size", len(expect.EntrySpans), len(got.EntrySpans))
	for i, gotEntrySpan := range got.EntrySpans {
		expectEntrySpan := expect.EntrySpans[i]
		checkSpan(t, fmt.Sprintf("EntrySpan %d", i), expectEntrySpan, gotEntrySpan)
	}
	checkIntEqual(t, "ExitSpans Size", len(expect.ExitSpans), len(got.ExitSpans))
	for i, gotExitSpan := range got.ExitSpans {
		expectExitSpan := expect.ExitSpans[i]
		checkSpan(t, fmt.Sprintf("ExitSpan %d", i), expectExitSpan, gotExitSpan)
	}
	checkIntEqual(t, "ErrorSpans Size", len(expect.ErrorSpans), len(got.ErrorSpans))
	for i, gotErrorSpan := range got.ErrorSpans {
		expectErrorSpan := expect.ErrorSpans[i]
		checkSpan(t, fmt.Sprintf("ErrorSpan %d", i), expectErrorSpan, gotErrorSpan)
	}

	checkIntEqual(t, "Service Children Size", len(expect.Children), len(got.Children))
	for i, gotChildService := range got.Children {
		expectChildService := expect.Children[i]
		checkServiceNode(t, expectChildService, gotChildService)
	}
}

func checkSpan(t *testing.T, name string, expect *model.OtelSpan, got *model.OtelSpan) {
	if expect == nil {
		t.Errorf("[Miss %s] ServiceName=%s", name, got.ServiceName)
		return
	}
	checkUint64Equal(t, "span.StartTime", expect.StartTime, got.StartTime)
	checkUint64Equal(t, "span.Duration", expect.Duration, got.Duration)
	checkStringEqual(t, "span.ServiceName", expect.ServiceName, got.ServiceName)
	checkStringEqual(t, "span.Name", expect.Name, got.Name)
	checkStringEqual(t, "span.SpanId", expect.SpanId, got.SpanId)
	checkStringEqual(t, "span.PSpanId", expect.PSpanId, got.PSpanId)
	checkStringEqual(t, "span.NextSpanId", expect.NextSpanId, got.NextSpanId)
	checkStringEqual(t, "span.Kind", expect.Kind.String(), got.Kind.String())
	checkStringEqual(t, "span.Code", expect.Code.String(), got.Code.String())
	checkIntEqual(t, "span.Attributes Size", len(expect.Attributes), len(got.Attributes))

	for k, v := range got.Attributes {
		if expVal, ok := expect.Attributes[k]; ok {
			checkStringEqual(t, fmt.Sprintf("span.Attributes[%s]", k), expVal, v)
		} else {
			t.Errorf("[Miss span.Attributes] %s=%s", k, v)
		}
	}
	checkIntEqual(t, "span.Exceptions Size", len(expect.Exceptions), len(got.Exceptions))
	for i, gotExcpetion := range got.Exceptions {
		expException := expect.Exceptions[i]
		checkUint64Equal(t, "exception.Timestamp", expException.Timestamp, gotExcpetion.Timestamp)
		checkStringEqual(t, "exception.Type", expException.Type, gotExcpetion.Type)
		checkStringEqual(t, "exception.Message", expException.Message, gotExcpetion.Message)
		checkStringEqual(t, "exception.Stack", expException.Stack, gotExcpetion.Stack)
	}
}

func isFileEmpty(f string) bool {
	fi, err := os.Stat(f)
	return err != nil || fi.Size() == 0
}

func checkStringEqual(t *testing.T, key string, expect string, got string) {
	if expect != got {
		t.Errorf("[Check %s] want=%s, got=%s", key, expect, got)
	}
}

func checkIntEqual(t *testing.T, key string, expect int, got int) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

func checkUint64Equal(t *testing.T, key string, expect uint64, got uint64) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

func convertSkywalkingToTraceCase(path string) (*TestTraceCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	response := &skywalking.SkywalkingResponse{}
	json.Unmarshal(data, response)

	serviceNodes, err := skywalking.ConvertToServiceNodes(&response.Data.Trace)
	if err != nil {
		return nil, err
	}
	return newTestTraceCase(response.Data.Trace.Spans[0].TraceId, serviceNodes), nil
}

func convertJaegerToTraceCase(path string) (*TestTraceCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	response := &jaeger.JaegerResponse{}
	json.Unmarshal(data, response)

	serviceNodes, err := jaeger.ConvertToServiceNodes(&response.Data[0])
	if err != nil {
		return nil, err
	}
	return newTestTraceCase(response.Data[0].TraceId, serviceNodes), nil
}

func convertESApmToTraceCase(path string) (*TestTraceCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	response := &elastic.SearchResp{}
	json.Unmarshal(data, response)

	serviceNodes, err := elastic.ConvertToServiceNodes(response)
	if err != nil {
		return nil, err
	}
	return newTestTraceCase(response.GetTraceId(), serviceNodes), nil
}

func convertPinpointToTraceCase(path string) (*TestTraceCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	response := &pinpoint.PinpointResponse{}
	json.Unmarshal(data, response)

	serviceNodes, err := response.ConvertToServiceNodes()
	if err != nil {
		return nil, err
	}
	return newTestTraceCase(response.TraceId, serviceNodes), nil
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err) == false
}

type TestTraceCase struct {
	Name     string                   `json:"name"`
	TraceId  string                   `json:"traceId"`
	Services []*model.OtelServiceNode `json:"services"`
}

func newTestTraceCase(traceId string, services []*model.OtelServiceNode) *TestTraceCase {
	return &TestTraceCase{
		TraceId:  traceId,
		Services: services,
	}
}

type TestTraceDetailCase struct {
	Spans []*model.OtelSpan `json:"spans"`
}
