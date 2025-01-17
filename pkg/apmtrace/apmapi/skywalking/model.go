package skywalking

import "encoding/json"

type SkywalkingResponse struct {
	Data SkywalkingData `json:"data"`
}

type SkywalkingData struct {
	Trace SkywalkingTrace `json:"trace"`
}

type SkywalkingTrace struct {
	Spans []*SkywalkingSpan `json:"spans"`
}

type SkywalkingSpan struct {
	TraceId      string                 `json:"traceId"`
	SegmentId    string                 `json:"segmentId"`
	SpanId       int                    `json:"spanId"`
	ParentSpanId int                    `json:"parentSpanId"`
	Refs         []*SkywalkingRef       `json:"refs"`
	ServiceCode  string                 `json:"serviceCode"`
	StartTime    uint64                 `json:"startTime"`
	EndTime      uint64                 `json:"endTime"`
	EndpointName string                 `json:"endpointName"`
	SpanType     SpanType               `json:"type"`
	Peer         string                 `json:"peer"`
	Component    string                 `json:"component"`
	IsError      bool                   `json:"isError"`
	SpanLayer    SpanLayer              `json:"layer"`
	Tags         []*SkywalkingKeyValue  `json:"tags"`
	Logs         []*SkywalkingLogEntity `json:"logs"`
}

func (s *SpanType) UnmarshalJSON(data []byte) error {
	var spanTypeStr string
	err := json.Unmarshal(data, &spanTypeStr)
	if err != nil {
		return err
	}

	if spanType, exist := SpanType_map[spanTypeStr]; exist {
		*s = spanType
	}
	return nil
}

func (s *SpanLayer) UnmarshalJSON(data []byte) error {
	var spanLayerStr string
	err := json.Unmarshal(data, &spanLayerStr)
	if err != nil {
		return err
	}

	if layer, exist := SpanLayer_map[spanLayerStr]; exist {
		*s = layer
	}
	return nil
}

type SkywalkingRef struct {
	TraceId         string  `json:"traceId"`
	ParentSegmentId string  `json:"parentSegmentId"`
	ParentSpanId    int     `json:"parentSpanId"`
	Type            RefType `json:"type"`
}

func (s *RefType) UnmarshalJSON(data []byte) error {
	var refTypeStr string
	err := json.Unmarshal(data, &refTypeStr)
	if err != nil {
		return err
	}

	if refType, exist := RefType_map[refTypeStr]; exist {
		*s = refType
	}
	return nil
}

type SkywalkingKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SkywalkingLogEntity struct {
	Time uint64                `json:"time"`
	Data []*SkywalkingKeyValue `json:"data"`
}
