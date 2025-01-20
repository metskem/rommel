package routes

import (
	"github.com/awesome-gocui/gocui"
	"time"
)

type RouteMetric struct {
	LastSeen      time.Time
	Host          string
	RTotal        int
	R2xx          int
	R3xx          int
	R4xx          int
	R5xx          int
	GETs          int
	POSTs         int
	PUTs          int
	DELETEs       int
	TotalRespTime int64
}

const (
	TagStatusCode = "status_code"
	TagUri        = "uri"
	TagMethod     = "method"
)

var (
	mainView       *gocui.View
	summaryView    *gocui.View
	RouteMetricMap = make(map[string]RouteMetric) // map key is app-guid
)
