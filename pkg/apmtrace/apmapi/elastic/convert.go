package elastic

import (
	"encoding/json"

	"github.com/CloudDetail/apo-module/apm/model/v1"
	cmodel "github.com/CloudDetail/apo-module/model/v1"
)

func ConvertToServiceNodes(resp *SearchResp) ([]*model.OtelServiceNode, error) {
	var otelSpans = []*model.OtelSpan{}
	var otelSpanMap = map[string]*model.OtelSpan{}
	for i := 0; i < len(resp.Hits.Hits); i++ {
		hit := resp.Hits.Hits[i]

		switch GetProcessEvent(&hit) {
		case SpanProcessor:
			otelSpan := rawSpanToOtelSpan(hit.Source)
			if otelSpan != nil {
				otelSpanMap[otelSpan.SpanId] = otelSpan
				otelSpans = append(otelSpans, otelSpan)
			}
		case TransactionProcessor:
			otelSpan := rawTransactionToOtelSpan(hit.Source)
			if otelSpan != nil {
				otelSpanMap[otelSpan.SpanId] = otelSpan
				otelSpans = append(otelSpans, otelSpan)
			}
		case ErrorProcessor:
			exceptions, parentId := rawErrorToException(hit.Source)
			if len(exceptions) == 0 {
				continue
			}
			if otelSpan, find := otelSpanMap[parentId]; find {
				otelSpan.Exceptions = append(otelSpan.Exceptions, exceptions...)
			}
		default:
			continue
		}
	}

	traceData := model.NewOTelTrace("elastic")
	if otelSpans == nil && len(otelSpans) == 0 {
		return traceData.GetServiceNodes(), nil
	}

	traceTree := model.NewOtelTree()
	for _, span := range otelSpans {
		if err := traceTree.AddSpan(span); err != nil {
			return nil, err
		}
	}
	if err := traceTree.BuildRelation4Spans(traceData); err != nil {
		return nil, err
	}

	return traceData.GetServiceNodes(), nil
}

func rawSpanToOtelSpan(source json.RawMessage) *model.OtelSpan {
	span := &Span{}
	err := json.Unmarshal(source, span)
	if err != nil {
		return nil
	}
	if span.Span.ID == "" {
		return nil
	}
	return convertSpanToOtelSpan(span)
}

func rawTransactionToOtelSpan(source json.RawMessage) *model.OtelSpan {
	transaction := &Transaction{}
	err := json.Unmarshal(source, transaction)
	if err != nil {
		return nil
	}
	if transaction.Transaction.ID == "" {
		return nil
	}
	return convertTransToOtelSpan(transaction)
}

func rawErrorToException(source json.RawMessage) (exceptions []*cmodel.Exception, parentId string) {
	errorSpan := &ErrorSpan{}
	err := json.Unmarshal(source, errorSpan)
	if err != nil {
		return nil, ""
	}
	if errorSpan.Parent.ID == "" {
		return nil, ""
	}

	return convertErrorToException(errorSpan)
}
