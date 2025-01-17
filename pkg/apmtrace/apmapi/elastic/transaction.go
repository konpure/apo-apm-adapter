package elastic

import (
	"strings"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

type Transaction struct {
	Trace     Trace     `json:"trace"`
	Parent    *Parent   `json:"parent"`
	Timestamp Timestamp `json:"timestamp"`
	Processor Processor `json:"processor"`
	Service   Service   `json:"service"`

	Transaction TransactionClass `json:"transaction"`

	URL   *URL  `json:"url"`
	HTTP  *HTTP `json:"http"`
	Event Event `json:"event"`

	// Ignore Fields.
	// Container            Container        `json:"container"`
	// Kubernetes           Kubernetes       `json:"kubernetes"`
	// Agent                Agent            `json:"agent"`
	// Source               Client           `json:"source"`
	// Observer             Observer         `json:"observer"`
	// Ecs                  Ecs              `json:"ecs"`
	// Host Host `json:"host"`
	// Client               Client           `json:"client"`

	// UserAgent            UserAgent        `json:"user_agent"`
	// Process   Process   `json:"process"`
	// TimestampStr   string           `json:"@timestamp"`
}

type Source struct {
	IP string `json:"ip"`
}

type Client struct {
	IP string `json:"ip"`
}

type Container struct {
	ID string `json:"id"`
}

type Trace struct {
	ID string `json:"id"`
}

type HTTP struct {
	Request  *Request  `json:"request"`
	Response *Response `json:"response"`
	Version  string    `json:"version"`
}

type OS struct {
	Platform string `json:"platform"`
}

type Kubernetes struct {
	Pod Pod `json:"pod"`
}

type Pod struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

type Process struct {
	PID   int64  `json:"pid"`
	Title string `json:"title"`
}

type Service struct {
	Name string `json:"name"`
	// Node      Node      `json:"node"`
	// Framework Framework `json:"framework"`
	// Runtime   Framework `json:"runtime"`
	// Language  Framework `json:"language"`
	// Version   string    `json:"version"`
}

type Framework struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Node struct {
	Name string `json:"name"`
}

type TransactionClass struct {
	Result    string    `json:"result"`
	Duration  Duration  `json:"duration"`
	Name      string    `json:"name"`
	SpanCount SpanCount `json:"span_count"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Sampled   bool      `json:"sampled"`
}

type Duration struct {
	Us int64 `json:"us"`
}

type SpanCount struct {
	Dropped int64 `json:"dropped"`
	Started int64 `json:"started"`
}

type URL struct {
	Path   string `json:"path"`
	Scheme string `json:"scheme"`
	Port   int64  `json:"port"`
	Query  string `json:"query"`
	Domain string `json:"domain"`
	Full   string `json:"full"`
}

type UserAgent struct {
	Original string `json:"original"`
	Name     string `json:"name"`
	Device   Device `json:"device"`
	Version  string `json:"version"`
}

type Device struct {
	Name string `json:"name"`
}

func (t *Transaction) ParentSpanId() string {
	if t.Parent == nil {
		return ""
	} else {
		return t.Parent.ID
	}
}

func (t *Transaction) Code() model.OtelStatusCode {
	if t.Event.Outcome == "success" {
		return model.StatusCodeOk
	}
	return model.StatusCodeError
}

func convertTransToOtelSpan(t *Transaction) *model.OtelSpan {
	if t == nil {
		return nil
	}

	entrySpan := &model.OtelSpan{
		StartTime:   uint64(t.Timestamp.Us) * 1e3,
		Duration:    uint64(t.Transaction.Duration.Us) * 1e3,
		ServiceName: t.Service.Name,
		Name:        t.Transaction.Name,
		SpanId:      t.Transaction.ID,
		PSpanId:     t.ParentSpanId(),
		// NextSpanId:  "",
		Kind:       model.SpanKindServer,
		Code:       t.Code(),
		Attributes: map[string]string{},
	}

	if t.URL != nil {
		entrySpan.Attributes[model.AttributeHTTPURL] = t.URL.Full
	}

	if strings.HasPrefix(t.Transaction.Result, "HTTP ") {
		entrySpan.Attributes[model.AttributeHTTPStatusCode] = t.Transaction.Result[5:]
	}

	return entrySpan
}
