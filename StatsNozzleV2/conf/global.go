package conf

import "sync"

const (
	MetricCpu            = "cpu"
	MetricAge            = "container_age"
	MetricCpuEntitlement = "cpu_entitlement"
	MetricDisk           = "disk"
	MetricMemory         = "memory"
	MetricLogRate        = "log_rate"
	TagOrgName           = "organization_name"
	TagSpaceName         = "space_name"
	TagAppName           = "app_name"
	ColorReset           = "\033[0m"
	ColorYellow          = "\033[33m"
	ColorBlue            = "\033[34m"
	ColorWhite           = "\033[97m"
)

var (
	MapLock        sync.Mutex
	MetricNames    = []string{MetricCpu, MetricAge, MetricCpuEntitlement, MetricDisk, MetricMemory, MetricLogRate}
	MetricMap      = make(map[string]Metric) // map key is app-guid/index
	TotalEnvelopes int
)

type Metric struct {
	AppIndex  string
	AppName   string
	SpaceName string
	OrgName   string
	CpuTot    float64
	Values    map[string]float64
}
