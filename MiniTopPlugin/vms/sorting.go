package vms

import (
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
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
)

var (
	upTimeColor                            = common.ColorWhite
	JobColor                               = common.ColorWhite
	containerUsageMemoryColor              = common.ColorWhite
	CapacityTotalDiskColor                 = common.ColorWhite
	containerUsageDiskColor                = common.ColorWhite
	containerCountColor                    = common.ColorWhite
	capacityTotalMemoryColor               = common.ColorWhite
	capacityAllocatedMemoryColor           = common.ColorWhite
	IPTablesRuleCountColor                 = common.ColorWhite
	NetInterfaceCountColor                 = common.ColorWhite
	OverlayTxBytesColor                    = common.ColorWhite
	OverlayRxBytesColor                    = common.ColorWhite
	OverlayRxDroppedColor                  = common.ColorWhite
	OverlayTxDropped                       = common.ColorWhite
	HTTPRouteCountColor                    = common.ColorWhite
	DopplerConnectionsColor                = common.ColorWhite
	ActiveDrainsColor                      = common.ColorWhite
	activeSortFieldColor         SortField = sortByJob
)

func spacePressed(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	common.FlipSortOrder()
	return nil
}

func colorSortedColumn() {
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
	case sortByNetInterfaceCount:
		NetInterfaceCountColor = common.ColorBlue
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
	case sortByDopplerConnections:
		DopplerConnectionsColor = common.ColorBlue
	case sortByActiveDrains:
		ActiveDrainsColor = common.ColorBlue
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
		return p[i].Value.UpTime < p[j].Value.UpTime
	case sortByContainerUsageMemory:
		return p[i].Value.ContainerUsageMemory < p[j].Value.ContainerUsageMemory
	case sortByCapacityTotalDisk:
		return p[i].Value.CapacityTotalDisk < p[j].Value.CapacityTotalDisk
	case sortByContainerUsageDisk:
		return p[i].Value.ContainerUsageDisk < p[j].Value.ContainerUsageDisk
	case sortByContainerCount:
		return p[i].Value.ContainerCount < p[j].Value.ContainerCount
	case sortByCapacityTotalMemory:
		return p[i].Value.CapacityTotalMemory < p[j].Value.CapacityTotalMemory
	case sortByCapacityAllocatedMemory:
		return p[i].Value.CapacityAllocatedMemory < p[j].Value.CapacityAllocatedMemory
	case sortByIPTablesRuleCount:
		return p[i].Value.IPTablesRuleCount < p[j].Value.IPTablesRuleCount
	case sortByNetInterfaceCount:
		return p[i].Value.NetInterfaceCount < p[j].Value.NetInterfaceCount
	case sortByOverlayTxBytes:
		return p[i].Value.OverlayTxBytes < p[j].Value.OverlayTxBytes
	case sortByOverlayRxBytes:
		return p[i].Value.OverlayRxBytes < p[j].Value.OverlayRxBytes
	case sortByOverlayRxDropped:
		return p[i].Value.OverlayRxDropped < p[j].Value.OverlayRxDropped
	case sortByOverlayTxDropped:
		return p[i].Value.OverlayTxDropped < p[j].Value.OverlayTxDropped
	case sortByHTTPRouteCount:
		return p[i].Value.HTTPRouteCount < p[j].Value.HTTPRouteCount
	case sortByDopplerConnections:
		return p[i].Value.DopplerConnections < p[j].Value.DopplerConnections
	case sortByActiveDrains:
		return p[i].Value.ActiveDrains < p[j].Value.ActiveDrains
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
	return filterPassed
}
