package conf

import (
	"sync"
	"time"
)

const (
	MetricCpu            = "cpu"
	MetricAge            = "container_age"
	MetricCpuEntitlement = "cpu_entitlement"
	MetricDisk           = "disk"
	MetricMemory         = "memory"
	MetricMemoryQuota    = "memory_quota"
	MetricLogRate        = "log_rate"
	MetricLogRateLimit   = "log_rate_limit"
	TagOrgName           = "organization_name"
	TagSpaceName         = "space_name"
	TagAppName           = "app_name"
	TagAppId             = "app_id"
	TagAppInstanceId     = "instance_id" // use this for app index
	TagOrigin            = "origin"
	TagOriginValueRep    = "rep"
	TagOriginValueRtr    = "gorouter"
	ColorReset           = "\033[0m"
	ColorYellow          = "\033[33m"
	ColorBlue            = "\033[34m"
	ColorWhite           = "\033[97m"
)

var (
	MapLock           sync.Mutex
	MetricNames       = []string{MetricCpu, MetricAge, MetricCpuEntitlement, MetricDisk, MetricMemory, MetricMemoryQuota, MetricLogRate, MetricLogRateLimit}
	MetricMap         = make(map[string]Metric) // map key is app-guid/index
	TotalEnvelopes    float64
	TotalEnvelopesRep float64
	TotalEnvelopesRtr float64
	TotalApps         = make(map[string]bool)
	ShowFilter        = false
	StartTime         = time.Now()
)

type Metric struct {
	LastSeen  time.Time
	AppIndex  string
	AppName   string
	SpaceName string
	OrgName   string
	CpuTot    float64
	LogRtr    float64
	LogRep    float64
	IP        string
	Values    map[string]float64
}
