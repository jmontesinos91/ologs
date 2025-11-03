package v2

import "sort"

// Values defines the structured data to be logged by the Logger
type Values map[string]interface{}

// Logger vars
const (
	TrackingID = "TrackingID"
	UserID     = "UserID"
	LayoutID   = "LayoutID"
	Role       = "Role"
	EventID    = "EventID"
	AlarmID    = "AlarmID"
	Latency    = "LatencyMS"
	Method     = "Method"
	Path       = "Path"
	IMEI       = "Imei"
)

// Logger ...
type Logger interface {
	Close()

	Debug(msg string, values ...Values)

	Error(msg string, values ...Values)

	Info(msg string, values ...Values)

	Warn(msg string, values ...Values)

	// WithValues returns a new logger with fixed key and values pairs
	WithValues(values Values) Logger
}

/******************************************************************************/
/* AUXILIARY FUNCTIONS                                                        */
/******************************************************************************/

func (v Values) toVariadic() []interface{} {
	keyAndValues := make([]interface{}, 0, len(v)*2)
	orderedSlice := []string{}
	for key := range v {
		orderedSlice = append(orderedSlice, key)
	}

	sort.Strings(orderedSlice)
	for _, key := range orderedSlice {
		keyAndValues = append(keyAndValues, key, v[key])
	}
	return keyAndValues
}
