package jaeger

type JaegerResponse struct {
	Data []JaegerData `json:"data"`
}

type JaegerData struct {
	TraceId   string                    `json:"traceId"`
	Spans     []*JaegerSpan             `json:"spans"`
	Processes map[string]*JaegerProcess `json:"processes"`
}

type JaegerSpan struct {
	TraceId       string            `json:"traceId"`
	SpanId        string            `json:"spanID"`
	OperationName string            `json:"operationName"`
	References    []*JaegerSpanRef  `json:"references"`
	StartTime     uint64            `json:"startTime"`
	Duration      uint64            `json:"duration"`
	Tags          []*JaegerKeyValue `json:"tags"`
	Logs          []*JaegerLog      `json:"logs"`
	ProcessID     string            `json:"processID"`
}

func (span *JaegerSpan) GetParentSpanID() string {
	if len(span.References) == 0 {
		return ""
	}
	refSpanId := ""
	for _, ref := range span.References {
		if ref.RefType == "CHILD_OF" {
			return ref.SpanID
		}
		if ref.RefType == "FOLLOWS_FROM" {
			refSpanId = ref.SpanID
		}
	}
	return refSpanId
}

type JaegerSpanRef struct {
	RefType string `json:"refType"`
	TraceId string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

type JaegerLog struct {
	Timestamp uint64            `json:"timestamp"`
	Fields    []*JaegerKeyValue `json:"fields"`
}

type JaegerKeyValue struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (kv *JaegerKeyValue) GetVStr() string {
	if value, ok := kv.Value.(string); ok {
		return value
	}
	return ""
}

func (kv *JaegerKeyValue) GetVBool() bool {
	if value, ok := kv.Value.(bool); ok {
		return value
	}
	return false
}

func (kv *JaegerKeyValue) GetVInt64() int64 {
	if value, ok := kv.Value.(float64); ok {
		return int64(value)
	}
	return 0
}

func (kv *JaegerKeyValue) GetVFloat64() float64 {
	if value, ok := kv.Value.(float64); ok {
		return value
	}
	return 0.0
}

func (kv *JaegerKeyValue) GetVBinary() []byte {
	if value, ok := kv.Value.([]byte); ok {
		return value
	}
	return nil
}

type JaegerProcess struct {
	ServiceName string            `json:"serviceName"`
	Tags        []*JaegerKeyValue `json:"tags"`
}
