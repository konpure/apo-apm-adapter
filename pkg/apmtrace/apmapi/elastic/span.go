package elastic

import (
	"strconv"

	"github.com/CloudDetail/apo-module/apm/model/v1"
	cmodel "github.com/CloudDetail/apo-module/model/v1"
)

type Span struct {
	Trace     Trace       `json:"trace"`
	Parent    Parent      `json:"parent"`
	Timestamp Timestamp   `json:"timestamp"`
	Processor Processor   `json:"processor"`
	Service   SpanService `json:"service"`

	Transaction SpanTransactionClass `json:"transaction"`
	Span        SpanClass            `json:"span"`

	URL         *SpanURL         `json:"url"`
	HTTP        *SpanHTTP        `json:"http"`
	Destination *SpanDestination `json:"destination"`
	Event       Event            `json:"event"`

	// Observer     Observer `json:"observer"`
	// Agent        Agent    `json:"agent"`
	// TimestampStr string   `json:"@timestamp"`
	// Ecs          Ecs      `json:"ecs"`
}

type SpanDestination struct {
	Address string `json:"address"`
	Port    int64  `json:"port"`
}

type SpanHTTP struct {
	SpanRequest  *SpanRequest  `json:"request"`
	SpanResponse *SpanResponse `json:"response"`
}

type SpanRequest struct {
	Method string `json:"method"`
}

type SpanResponse struct {
	StatusCode int `json:"status_code"`
}

type SpanService struct {
	Name string `json:"name"`
}

type SpanClass struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Subtype  string   `json:"subtype"`
	Duration Duration `json:"duration"`

	Destination *SpanDestinationClass `json:"destination"`
	HTTP        *SpanHTTPClass        `json:"http"`
	DB          *SpanDBClass          `json:"db"`

	// Stacktrace  []Stacktrace         `json:"stacktrace"`
	HTTPURLOriginal string `json:"http.url.original"`
}

type SpanDestinationClass struct {
	Service DestinationService `json:"service"`
}

type DestinationService struct {
	Resource string `json:"resource"`
}

type SpanHTTPClass struct {
	Method   string       `json:"method"`
	Response SpanResponse `json:"response"`
}

type SpanDBClass struct {
	Instance  string `json:"instance"`
	Statement string `json:"statement"`
	Type      string `json:"type"`
}

type SpanURL struct {
	Original string `json:"original"`
}

type SpanTransactionClass struct {
	Id string `json:"id"`
}

func (span *Span) Code() model.OtelStatusCode {
	if span.Event.Outcome == "success" {
		return model.StatusCodeOk
	}
	return model.StatusCodeError
}

func (span *Span) SpanKind() model.OtelSpanKind {
	if span.Span.Destination != nil {
		return model.SpanKindClient
	}

	return model.SpanKindInternal
}

func convertSpanToOtelSpan(span *Span) *model.OtelSpan {
	if span == nil {
		return nil
	}

	otelSpan := &model.OtelSpan{
		StartTime:   uint64(span.Timestamp.Us) * 1e3,
		Duration:    uint64(span.Span.Duration.Us) * 1e3,
		ServiceName: span.Service.Name,
		Name:        span.Span.Name,
		SpanId:      span.Span.ID,
		PSpanId:     span.Parent.ID,
		NextSpanId:  "",
		Kind:        span.SpanKind(),
		Code:        span.Code(),
		Attributes:  map[string]string{},
		Exceptions:  []*cmodel.Exception{},
	}

	if span.Span.DB != nil {
		otelSpan.Attributes[model.AttributeDBName] = span.Span.DB.Instance
		otelSpan.Attributes[model.AttributeDBStatement] = span.Span.DB.Statement
		otelSpan.Attributes[model.AttributeDBSystem] = span.Span.DB.Type
	}

	if span.Span.HTTP != nil {
		otelSpan.Attributes[model.AttributeHTTPURL] = span.Span.HTTPURLOriginal
		otelSpan.Attributes[model.AttributeHTTPStatusCode] = strconv.Itoa(span.Span.HTTP.Response.StatusCode)
	}
	return otelSpan
}
