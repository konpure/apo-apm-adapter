package jaeger

import "github.com/CloudDetail/apo-module/apm/model/v1"

// Some of the keys used to represent OTLP constructs as tags or annotations in other formats.
const (
	TagSpanKind = "span.kind"

	TagError         = "error"
	TagHTTPStatusMsg = "http.status_message"

	TagW3CTraceState = "w3c.tracestate"

	OtelStatusCode = "otel.status_code"
)

// Constants used for signifying batch-level attribute values where not supplied by OTLP data but required
// by other protocols.
const (
	ResourceNoServiceName = "OTLPResourceNoServiceName"
)

// OpenTracingSpanKind are possible values for TagSpanKind and match the OpenTracing
// conventions: https://github.com/opentracing/specification/blob/main/semantic_conventions.md
// These values are used for representing span kinds that have no
// equivalents in OpenCensus format. They are stored as values of TagSpanKind
type OpenTracingSpanKind string

const (
	OpenTracingSpanKindUnspecified OpenTracingSpanKind = ""
	OpenTracingSpanKindClient      OpenTracingSpanKind = "client"
	OpenTracingSpanKindServer      OpenTracingSpanKind = "server"
	OpenTracingSpanKindConsumer    OpenTracingSpanKind = "consumer"
	OpenTracingSpanKindProducer    OpenTracingSpanKind = "producer"
	OpenTracingSpanKindInternal    OpenTracingSpanKind = "internal"
)

// StatusCodeFromHTTP takes an HTTP status code and return the appropriate OpenTelemetry status code
// See: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/http.md#status
func StatusCodeFromHTTP(httpStatusCode int64) model.OtelStatusCode {
	if httpStatusCode >= 100 && httpStatusCode < 399 {
		return model.StatusCodeUnset
	}
	return model.StatusCodeError
}
