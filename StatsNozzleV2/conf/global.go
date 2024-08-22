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
	ColorRed             = "\033[31m"
	ColorGreen           = "\033[32m"
	ColorGreenBright     = "\033[32m;1m"
	ColorYellow          = "\033[33m"
	ColorYellowBright    = "\033[33;1m"
	ColorBlue            = "\033[34m"
	ColorMagenta         = "\033[35m"
	ColorCyan            = "\033[36m"
	ColorGray            = "\033[37m"
	ColorWhite           = "\033[97m"
	ColorWhiteBright     = "\033[97m;1m"
)

var (
	MapLock     sync.Mutex
	MetricNames = []string{MetricCpu, MetricAge, MetricCpuEntitlement, MetricDisk, MetricMemory, MetricLogRate}
	//MetricMap   = make(map[string]map[string]float64) // key: org/space/app/index, value: map of metricName to value
	MetricMap = make(map[string]Metric) // map key is app-guid/index
)

type Metric struct {
	AppIndex  string
	AppName   string
	SpaceName string
	OrgName   string
	Values    map[string]float64
}
