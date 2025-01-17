package skywalking

// Map to the type of span
type SpanType int32

const (
	// Server side of RPC. Consumer side of MQ.
	SpanType_Entry SpanType = 0
	// Client side of RPC. Producer side of MQ.
	SpanType_Exit SpanType = 1
	// A common local code execution.
	SpanType_Local SpanType = 2
)

// Enum value maps for SpanType.
var (
	SpanType_map = map[string]SpanType{
		"Entry": SpanType_Entry,
		"Exit":  SpanType_Exit,
		"Local": SpanType_Local,
	}
)

// Type of the reference
type RefType int32

const (
	// Map to the reference targeting the segment in another OS process.
	RefType_CrossProcess RefType = 0
	// Map to the reference targeting the segment in the same process of the current one, just across thread.
	// This is only used when the coding language has the thread concept.
	RefType_CrossThread RefType = 1
)

// Enum value maps for RefType.
var (
	RefType_name = map[int32]string{
		0: "CrossProcess",
		1: "CrossThread",
	}

	RefType_map = map[string]RefType{
		"CrossProcess": RefType_CrossProcess,
		"CrossThread":  RefType_CrossThread,
	}
)

func (x RefType) String() string {
	return RefType_name[int32(x)]
}

// Map to the layer of span
type SpanLayer int32

const (
	// Unknown layer. Could be anything.
	SpanLayer_Unknown SpanLayer = 0
	// A database layer, used in tracing the database client component.
	SpanLayer_Database SpanLayer = 1
	// A RPC layer, used in both client and server sides of RPC component.
	SpanLayer_RPCFramework SpanLayer = 2
	// HTTP is a more specific RPCFramework.
	SpanLayer_Http SpanLayer = 3
	// A MQ layer, used in both producer and consuer sides of the MQ component.
	SpanLayer_MQ SpanLayer = 4
	// A cache layer, used in tracing the cache client component.
	SpanLayer_Cache SpanLayer = 5
	// A FAAS layer, used in function-as-a-Service platform.
	SpanLayer_FAAS SpanLayer = 6
)

// Enum value maps for SpanLayer.
var (
	SpanLayer_map = map[string]SpanLayer{
		"Unknown":      SpanLayer_Unknown,
		"Database":     SpanLayer_Database,
		"RPCFramework": SpanLayer_RPCFramework,
		"Http":         SpanLayer_Http,
		"MQ":           SpanLayer_MQ,
		"Cache":        SpanLayer_Cache,
		"FAAS":         SpanLayer_FAAS,
	}
)
