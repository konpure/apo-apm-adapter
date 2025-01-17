package elastic

type Agent struct {
	Name        string `json:"name"`
	EphemeralID string `json:"ephemeral_id"`
	Version     string `json:"version"`
}

type Parent struct {
	ID string `json:"id"`
}

type Ecs struct {
	Version string `json:"version"`
}

type Event struct {
	Ingested string `json:"ingested"`
	Outcome  string `json:"outcome"`
}

type Request struct {
	// Headers RequestHeaders `json:"headers"`
	Method string `json:"method"`
}

type RequestHeaders map[string][]string

type Response struct {
	// Headers     ResponseHeaders `json:"headers"`
	StatusCode  int64 `json:"status_code"`
	Finished    bool  `json:"finished"`
	HeadersSent bool  `json:"headers_sent"`
}

type ResponseHeaders map[string][]string

type Observer struct {
	Hostname     string `json:"hostname"`
	ID           string `json:"id"`
	EphemeralID  string `json:"ephemeral_id"`
	Type         string `json:"type"`
	Version      string `json:"version"`
	VersionMajor int64  `json:"version_major"`
}

type Processor struct {
	Name  string `json:"name"`
	Event string `json:"event"`
}

type Timestamp struct {
	Us int64 `json:"us"`
}

type Stacktrace struct {
	LibraryFrame        bool   `json:"library_frame"`
	ExcludeFromGrouping bool   `json:"exclude_from_grouping"`
	Filename            string `json:"filename"`
	Classname           string `json:"classname"`
	Line                Line   `json:"line"`
	Module              string `json:"module"`
	Function            string `json:"function"`
}

type Line struct {
	Number int64 `json:"number"`
}

type Host struct {
	OS           OS     `json:"os"`
	IP           string `json:"ip"`
	Architecture string `json:"architecture"`
}
