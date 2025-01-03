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
	ColorBlue            = "\033[36m"
	ColorWhite           = "\033[97m"

	AppOrInstanceViewInstance int = iota
	AppOrInstanceViewApp
)
const (
	FilterFieldAppName int = iota
	FilterFieldOrg
	FilterFieldSpace
)

var (
	MapLock                 sync.Mutex
	MetricNames             = []string{MetricCpu, MetricAge, MetricCpuEntitlement, MetricDisk, MetricMemory, MetricMemoryQuota, MetricLogRate, MetricLogRateLimit}
	InstanceMetricMap       = make(map[string]AppOrInstanceMetric) // map key is app-guid/index
	AppMetricMap            map[string]AppOrInstanceMetric         // map key is app-guid
	TotalEnvelopes          float64
	TotalEnvelopesPerSec    float64
	TotalEnvelopesRep       float64
	TotalEnvelopesRepPerSec float64
	TotalEnvelopesRtr       float64
	TotalEnvelopesRtrPerSec float64
	TotalApps               = make(map[string]bool)
	TotalMemoryUsed         float64
	TotalMemoryAllocated    float64
	TotalLogRateUsed        float64
	AppInstanceCounters     = make(map[string]AppInstanceCounter) // here we keep the highest instance index for each app
	ShowFilter              = false
	ShowHelp                = false
	StartTime               = time.Now()
	FilterStrings           = make(map[int]string)
	IntervalSecs            = 1
	AppOrInstanceView       = AppOrInstanceViewInstance
)

type AppInstanceCounter struct {
	Count       int
	LastUpdated time.Time
}

type AppOrInstanceMetric struct {
	LastSeen  time.Time
	AppIndex  string
	IxCount   int
	AppName   string
	AppGuid   string
	SpaceName string
	OrgName   string
	CpuTot    float64
	LogRtr    float64
	LogRep    float64
	IP        string
	Tags      map[string]float64
}
