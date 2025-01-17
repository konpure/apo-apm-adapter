package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/tidwall/gjson"
)

var droppedFields = []string{
	"span.stacktrace",
	"observer",
	"kubernetes",
	"container",
	"agent",
	"process",
	"http.headers",
	"response.headers",
	"user_agent",
}

type ESClient struct {
	es *elasticsearch.Client
}

type SearchResp struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []UnpackerHit `json:"hits"`
	} `json:"hits"`
}

func (r *SearchResp) GetTraceId() string {
	if len(r.Hits.Hits) > 0 {
		source := r.Hits.Hits[0].Source
		return gjson.Get(string(source), "trace.id").String()
	}
	return ""
}

type UnpackerHit struct {
	Index  string              `json:"_index"`
	Source json.RawMessage     `json:"_source"`
	Fields map[string][]string `json:"fields"`
}

type ProcessorEvent string

const (
	SpanProcessor        ProcessorEvent = "span"
	TransactionProcessor ProcessorEvent = "transaction"
	ErrorProcessor       ProcessorEvent = "error"
	UnknownProcessor     ProcessorEvent = "unknown"
)

func GetProcessEvent(hit *UnpackerHit) ProcessorEvent {
	if hit.Fields != nil {
		if len(hit.Fields["processor.event"]) > 0 {
			return ProcessorEvent(hit.Fields["processor.event"][0])
		}
	}
	return UnknownProcessor
}

func (c *ESClient) searchSpans(traceId string, indices ...string) (*SearchResp, error) {
	var buf bytes.Buffer
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{
				"trace.id": traceId,
			},
		},
		"_source": map[string]any{
			"excludes": droppedFields,
		},
		"fields": []string{"processor.event"},
		"size":   200,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %s", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(context.Background()),
		c.es.Search.WithIndex(indices...),
		c.es.Search.WithBody(&buf),
	)

	if err != nil {
		return nil, fmt.Errorf("search query error: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search query error: %s", res.String())
	}

	var result = SearchResp{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}

	return &result, nil
}
