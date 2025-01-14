package vms

import (
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"regexp"
	"sort"
)

const (
	sortByLastSeen = iota
	sortByAge
	sortByIP
	sortByJob
	sortByUpTime
	sortByContainerUsageMemory
	sortByCapacityTotalDisk
	sortByContainerUsageDisk
	sortByContainerCount
	sortByCapacityTotalMemory
	sortByCapacityAllocatedMemory
	sortByIPTablesRuleCount
	sortByNetInterfaceCount
	sortByOverlayTxBytes
	sortByOverlayRxBytes
	sortByOverlayRxDropped
	sortByOverlayTxDropped
	sortByHTTPRouteCount
	sortByDopplerConnections
	sortByActiveDrains
	sortByNumCPUS
	sortByResponses
	sortBy2xx
	sortBy3xx
	sortBy4xx
	sortBy5xx
)

var (
	upTimeColor                  = common.ColorWhite
	JobColor                     = common.ColorWhite
	containerUsageMemoryColor    = common.ColorWhite
	CapacityTotalDiskColor       = common.ColorWhite
	containerUsageDiskColor      = common.ColorWhite
	containerCountColor          = common.ColorWhite
	capacityTotalMemoryColor     = common.ColorWhite
	capacityAllocatedMemoryColor = common.ColorWhite
	IPTablesRuleCountColor       = common.ColorWhite
	//NetInterfaceCountColor                 = common.ColorWhite
	OverlayTxBytesColor   = common.ColorWhite
	OverlayRxBytesColor   = common.ColorWhite
	OverlayRxDroppedColor = common.ColorWhite
	OverlayTxDropped      = common.ColorWhite
	HTTPRouteCountColor   = common.ColorWhite
	//DopplerConnectionsColor                = common.ColorWhite
	//ActiveDrainsColor                      = common.ColorWhite
	numCPUSColor                   = common.ColorWhite
	responsesColor                 = common.ColorWhite
	r2xxColor                      = common.ColorWhite
	r3xxColor                      = common.ColorWhite
	r4xxColor                      = common.ColorWhite
	r5xxColor                      = common.ColorWhite
	activeSortFieldColor SortField = sortByIP
)

func spacePressed(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	common.FlipSortOrder()
	return nil
}

func colorSortedColumn() {
	util.WriteToFileDebug("colorSortedColumn VMs")
	common.LastSeenColor = common.ColorWhite
	common.AgeColor = common.ColorWhite
	//ixColor = common.ColorWhite
	common.IPColor = common.ColorWhite
	switch activeSortFieldColor {
	case sortByLastSeen:
		common.LastSeenColor = common.ColorBlue
	case sortByAge:
		common.AgeColor = common.ColorBlue
	case sortByJob:
		JobColor = common.ColorBlue
	case sortByIP:
		common.IPColor = common.ColorBlue
	case sortByUpTime:
		upTimeColor = common.ColorBlue
	case sortByContainerUsageMemory:
		containerUsageMemoryColor = common.ColorBlue
	case sortByCapacityTotalDisk:
		CapacityTotalDiskColor = common.ColorBlue
	case sortByContainerUsageDisk:
		containerUsageDiskColor = common.ColorBlue
	case sortByContainerCount:
		containerCountColor = common.ColorBlue
	case sortByCapacityTotalMemory:
		capacityTotalMemoryColor = common.ColorBlue
	case sortByCapacityAllocatedMemory:
		capacityAllocatedMemoryColor = common.ColorBlue
	case sortByIPTablesRuleCount:
		IPTablesRuleCountColor = common.ColorBlue
	//case sortByNetInterfaceCount:
	//	NetInterfaceCountColor = common.ColorBlue
	case sortByOverlayTxBytes:
		OverlayTxBytesColor = common.ColorBlue
	case sortByOverlayRxBytes:
		OverlayRxBytesColor = common.ColorBlue
	case sortByOverlayRxDropped:
		OverlayRxDroppedColor = common.ColorBlue
	case sortByOverlayTxDropped:
		OverlayTxDropped = common.ColorBlue
	case sortByHTTPRouteCount:
		HTTPRouteCountColor = common.ColorBlue
	//case sortByDopplerConnections:
	//	DopplerConnectionsColor = common.ColorBlue
	//case sortByActiveDrains:
	//	ActiveDrainsColor = common.ColorBlue
	case sortByNumCPUS:
		numCPUSColor = common.ColorBlue
	case sortByResponses:
		responsesColor = common.ColorBlue
	case sortBy2xx:
		r2xxColor = common.ColorBlue
	case sortBy3xx:
		r3xxColor = common.ColorBlue
	case sortBy4xx:
		r4xxColor = common.ColorBlue
	case sortBy5xx:
		r5xxColor = common.ColorBlue
	}
}

// based on https://stackoverflow.com/questions/18695346/how-to-sort-a-mapstringint-by-its-values
type SortField int

func sortedBy(metricMap map[string]CellMetric, reverse bool, sortField SortField) PairList {
	pairList := make(PairList, len(metricMap))
	i := 0
	for k, v := range metricMap {
		pairList[i] = Pair{sortField, k, v}
		i++
	}
	if reverse {
		sort.Sort(sort.Reverse(pairList))
	} else {
		sort.Sort(pairList)
	}
	return pairList
}

type PairList []Pair
type Pair struct {
	SortBy SortField
	Key    string
	Value  CellMetric
}

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool {
	switch p[i].SortBy {
	case sortByLastSeen:
		return p[i].Value.LastSeen.Unix() < p[j].Value.LastSeen.Unix()
	case sortByAge:
		return p[i].Value.Tags[metricAge] < p[j].Value.Tags[metricAge]
	case sortByJob:
		return p[i].Value.Job < p[j].Value.Job
	case sortByIP:
		return p[i].Value.IP < p[j].Value.IP
	case sortByUpTime:
		return p[i].Value.Tags[metricUpTime] < p[j].Value.Tags[metricUpTime]
	case sortByContainerUsageMemory:
		return p[i].Value.Tags[metricContainerUsageMemory] < p[j].Value.Tags[metricContainerUsageMemory]
	case sortByCapacityTotalDisk:
		return p[i].Value.Tags[metricCapacityTotalDisk] < p[j].Value.Tags[metricCapacityTotalDisk]
	case sortByContainerUsageDisk:
		return p[i].Value.Tags[metricContainerUsageDisk] < p[j].Value.Tags[metricContainerUsageDisk]
	case sortByContainerCount:
		return p[i].Value.Tags[metricContainerCount] < p[j].Value.Tags[metricContainerCount]
	case sortByCapacityTotalMemory:
		return p[i].Value.Tags[metricCapacityTotalMemory] < p[j].Value.Tags[metricCapacityTotalMemory]
	case sortByCapacityAllocatedMemory:
		return p[i].Value.Tags[metricCapacityAllocatedMemory] < p[j].Value.Tags[metricCapacityAllocatedMemory]
	case sortByIPTablesRuleCount:
		return p[i].Value.Tags[metricIPTablesRuleCount] < p[j].Value.Tags[metricIPTablesRuleCount]
	//case sortByNetInterfaceCount:
	//	return p[i].Value.Tags[metricNetInterfaceCount] < p[j].Value.Tags[metricNetInterfaceCount]
	case sortByOverlayTxBytes:
		return p[i].Value.Tags[metricOverlayTxBytes] < p[j].Value.Tags[metricOverlayTxBytes]
	case sortByOverlayRxBytes:
		return p[i].Value.Tags[metricOverlayRxBytes] < p[j].Value.Tags[metricOverlayRxBytes]
	case sortByOverlayRxDropped:
		return p[i].Value.Tags[metricOverlayRxDropped] < p[j].Value.Tags[metricOverlayRxDropped]
	case sortByOverlayTxDropped:
		return p[i].Value.Tags[metricOverlayTxDropped] < p[j].Value.Tags[metricOverlayTxDropped]
	case sortByHTTPRouteCount:
		return p[i].Value.Tags[metricHTTPRouteCount] < p[j].Value.Tags[metricHTTPRouteCount]
	case sortByDopplerConnections:
		return p[i].Value.Tags[metricDopplerConnections] < p[j].Value.Tags[metricDopplerConnections]
	case sortByActiveDrains:
		return p[i].Value.Tags[metricActiveDrains] < p[j].Value.Tags[metricActiveDrains]
	case sortByNumCPUS:
		return p[i].Value.Tags[metricNumCPUS] < p[j].Value.Tags[metricNumCPUS]
	case sortByResponses:
		return p[i].Value.Tags[metricResponses] < p[j].Value.Tags[metricResponses]
	case sortBy2xx:
		return p[i].Value.Tags[metric2xx] < p[j].Value.Tags[metric2xx]
	case sortBy3xx:
		return p[i].Value.Tags[metric3xx] < p[j].Value.Tags[metric3xx]
	case sortBy4xx:
		return p[i].Value.Tags[metric4xx] < p[j].Value.Tags[metric4xx]
	case sortBy5xx:
		return p[i].Value.Tags[metric5xx] < p[j].Value.Tags[metric5xx]
	}
	return p[i].Value.Tags[metricAge] > p[j].Value.Tags[metricAge] // default
}
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func passFilter(pairList Pair) bool {
	filterPassed := true
	filterRegex := regexp.MustCompile(common.FilterStrings[filterFieldIP])
	if !(common.FilterStrings[filterFieldIP] == "") && !filterRegex.MatchString(pairList.Value.IP) {
		filterPassed = false
	}
	oneTagValueFound := false
	for _, value := range pairList.Value.Tags {
		if value > 0 {
			oneTagValueFound = true
			break
		}
	}
	if oneTagValueFound {
		return filterPassed
	}
	return false
}
