package elastic

import (
	"strconv"
	"strings"

	cmodel "github.com/CloudDetail/apo-module/model/v1"
)

type ErrorSpan struct {
	Trace       Trace            `json:"trace"`
	Parent      Parent           `json:"parent"`
	Process     Process          `json:"process"`
	Error       Error            `json:"error"`
	Message     string           `json:"message"`
	Processor   Processor        `json:"processor"`
	Service     Service          `json:"service"`
	Transaction ErrorTransaction `json:"transaction"`
	Timestamp   Timestamp        `json:"timestamp"`

	// Agent Agent `json:"agent"`
	// Container  Container  `json:"container"`
	// Kubernetes Kubernetes `json:"kubernetes"`
	// Observer  Observer  `json:"observer"`
	// Ecs           Ecs              `json:"ecs"`
	// TimestampStr     string           `json:"@timestamp"`
	// Host        Host             `json:"host"`
	// Event       Event            `json:"event"`
}

type Error struct {
	Exception    []Exception `json:"exception"`
	ID           string      `json:"id"`
	GroupingKey  string      `json:"grouping_key"`
	GroupingName string      `json:"grouping_name"`
}

type Exception struct {
	Stacktrace []Stacktrace `json:"stacktrace"`
	Message    string       `json:"message"`
	Type       string       `json:"type"`
}

func (e *Exception) GetStacktrace() string {
	var strBuf strings.Builder

	strBuf.WriteString(e.Message)
	strBuf.WriteByte('\n')
	for i := 0; i < len(e.Stacktrace); i++ {
		strBuf.WriteString("  at ")
		strBuf.WriteString(e.Stacktrace[i].Classname)
		strBuf.WriteByte('.')
		strBuf.WriteString(e.Stacktrace[i].Function)
		strBuf.WriteByte('(')
		strBuf.WriteString(e.Stacktrace[i].Filename)
		strBuf.WriteByte(':')
		strBuf.WriteString(strconv.Itoa(int(e.Stacktrace[i].Line.Number)))
		strBuf.WriteByte(')')
		strBuf.WriteString("\n")
	}

	return strBuf.String()
}

type Language struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ErrorTransaction struct {
	ID      string `json:"id"`
	Sampled bool   `json:"sampled"`
}

func convertErrorToException(errorSpan *ErrorSpan) ([]*cmodel.Exception, string) {
	if errorSpan == nil {
		return nil, ""
	}

	var exceptions = []*cmodel.Exception{}
	for i := 0; i < len(errorSpan.Error.Exception); i++ {
		exception := &cmodel.Exception{
			Timestamp: uint64(errorSpan.Timestamp.Us),
			Type:      errorSpan.Error.Exception[i].Type,
			Message:   errorSpan.Error.Exception[i].Message,
			Stack:     errorSpan.Error.Exception[i].GetStacktrace(),
		}
		exceptions = append(exceptions, exception)
	}
	return exceptions, errorSpan.Parent.ID
}
