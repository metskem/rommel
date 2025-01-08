package apps

import (
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"regexp"
	"sort"
)

var (
	appNameColor                       = common.ColorWhite
	cpuPercColor                       = common.ColorWhite
	ixColor                            = common.ColorWhite
	cpuTotColor                        = common.ColorWhite
	memoryColor                        = common.ColorWhite
	memoryLimitColor                   = common.ColorWhite
	diskColor                          = common.ColorWhite
	logRateColor                       = common.ColorWhite
	logRateLimitColor                  = common.ColorWhite
	entColor                           = common.ColorWhite
	logRepColor                        = common.ColorWhite
	logRtrColor                        = common.ColorWhite
	orgColor                           = common.ColorWhite
	spaceColor                         = common.ColorWhite
	activeInstancesSortField SortField = sortByCpuPerc
	activeAppsSortField      SortField = sortByCpuPerc
)

func arrowRight(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	if activeInstancesSortField != sortBySpace {
		activeInstancesSortField++
	}
	if activeAppsSortField != sortBySpace {
		activeAppsSortField++
	}
	// when in instance view mode, there is no Ix column, so skip it
	if common.ActiveView == common.AppInstanceView {
		if activeInstancesSortField == sortByIx {
			activeInstancesSortField++
		}
	}
	// when in app view mode, the Age and IP columns are not there, so skip them
	if common.ActiveView == common.AppView {
		if activeAppsSortField == sortByAge || activeAppsSortField == sortByIP {
			activeAppsSortField++
		}
	}
	colorSortedColumn()
	return nil
}

func arrowLeft(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	if activeInstancesSortField != sortByAppName {
		activeInstancesSortField--
	}
	if activeAppsSortField != sortByAppName {
		activeAppsSortField--
	}
	// when in instance view mode, there is no Ix column, so skip it
	if common.ActiveView == common.AppInstanceView {
		if activeInstancesSortField == sortByIx {
			activeInstancesSortField--
		}
	}
	// when in app view mode, the Age and IP columns are not there, so skip them
	if common.ActiveView == common.AppView {
		if activeAppsSortField == sortByAge || activeAppsSortField == sortByIP {
			activeAppsSortField--
		}
	}
	colorSortedColumn()
	return nil
}

func colorSortedColumn() {
	appNameColor = common.ColorWhite
	common.LastSeenColor = common.ColorWhite
	common.AgeColor = common.ColorWhite
	cpuPercColor = common.ColorWhite
	ixColor = common.ColorWhite
	cpuTotColor = common.ColorWhite
	memoryColor = common.ColorWhite
	memoryLimitColor = common.ColorWhite
	diskColor = common.ColorWhite
	logRateColor = common.ColorWhite
	logRateLimitColor = common.ColorWhite
	entColor = common.ColorWhite
	common.IPColor = common.ColorWhite
	logRepColor = common.ColorWhite
	logRtrColor = common.ColorWhite
	orgColor = common.ColorWhite
	spaceColor = common.ColorWhite
	if common.ActiveView == common.AppInstanceView {
		switch activeInstancesSortField {
		case sortByAppName:
			appNameColor = common.ColorBlue
		case sortByLastSeen:
			common.LastSeenColor = common.ColorBlue
		case sortByAge:
			common.AgeColor = common.ColorBlue
		case sortByIx:
			ixColor = common.ColorBlue
		case sortByCpuPerc:
			cpuPercColor = common.ColorBlue
		case sortByCpuTot:
			cpuTotColor = common.ColorBlue
		case sortByMemory:
			memoryColor = common.ColorBlue
		case sortByMemoryLimit:
			memoryLimitColor = common.ColorBlue
		case sortByDisk:
			diskColor = common.ColorBlue
		case sortByLogRate:
			logRateColor = common.ColorBlue
		case sortByLogRateLimit:
			logRateLimitColor = common.ColorBlue
		case sortByIP:
			common.IPColor = common.ColorBlue
		case sortByEntitlement:
			entColor = common.ColorBlue
		case sortByLogRep:
			logRepColor = common.ColorBlue
		case sortByLogRtr:
			logRtrColor = common.ColorBlue
		case sortByOrg:
			orgColor = common.ColorBlue
		case sortBySpace:
			spaceColor = common.ColorBlue
		}
	}
	if common.ActiveView == common.AppView {
		switch activeAppsSortField {
		case sortByAppName:
			appNameColor = common.ColorBlue
		case sortByLastSeen:
			common.LastSeenColor = common.ColorBlue
		case sortByAge:
			common.AgeColor = common.ColorBlue
		case sortByIx:
			ixColor = common.ColorBlue
		case sortByCpuPerc:
			cpuPercColor = common.ColorBlue
		case sortByCpuTot:
			cpuTotColor = common.ColorBlue
		case sortByMemory:
			memoryColor = common.ColorBlue
		case sortByMemoryLimit:
			memoryLimitColor = common.ColorBlue
		case sortByDisk:
			diskColor = common.ColorBlue
		case sortByLogRate:
			logRateColor = common.ColorBlue
		case sortByLogRateLimit:
			logRateLimitColor = common.ColorBlue
		case sortByIP:
			common.IPColor = common.ColorBlue
		case sortByEntitlement:
			entColor = common.ColorBlue
		case sortByLogRep:
			logRepColor = common.ColorBlue
		case sortByLogRtr:
			logRtrColor = common.ColorBlue
		case sortByOrg:
			orgColor = common.ColorBlue
		case sortBySpace:
			spaceColor = common.ColorBlue
		}
	}
}

// based on https://stackoverflow.com/questions/18695346/how-to-sort-a-mapstringint-by-its-values
type SortField int

const (
	sortByAppName = iota
	sortByLastSeen
	sortByAge
	sortByIx
	sortByCpuPerc
	sortByCpuTot
	sortByMemory
	sortByMemoryLimit
	sortByDisk
	sortByLogRate
	sortByLogRateLimit
	sortByEntitlement
	sortByIP
	sortByLogRep
	sortByLogRtr
	sortByOrg
	sortBySpace
)

func sortedBy(metricMap map[string]AppOrInstanceMetric, reverse bool, sortField SortField) PairList {
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
	Value  AppOrInstanceMetric
}

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool {
	switch p[i].SortBy {
	case sortByAppName:
		return p[i].Value.AppName < p[j].Value.AppName
	case sortByLastSeen:
		return p[i].Value.LastSeen.Unix() < p[j].Value.LastSeen.Unix()
	case sortByAge:
		return p[i].Value.Tags[metricAge] < p[j].Value.Tags[metricAge]
	case sortByCpuPerc:
		return p[i].Value.Tags[MetricCpu] < p[j].Value.Tags[MetricCpu]
	case sortByCpuTot:
		return p[i].Value.CpuTot < p[j].Value.CpuTot
	case sortByMemory:
		return p[i].Value.Tags[metricMemory] < p[j].Value.Tags[metricMemory]
	case sortByMemoryLimit:
		return p[i].Value.Tags[metricMemoryQuota] < p[j].Value.Tags[metricMemoryQuota]
	case sortByDisk:
		return p[i].Value.Tags[metricDisk] < p[j].Value.Tags[metricDisk]
	case sortByEntitlement:
		return p[i].Value.Tags[metricCpuEntitlement] < p[j].Value.Tags[metricCpuEntitlement]
	case sortByIP:
		return p[i].Value.IP < p[j].Value.IP
	case sortByLogRate:
		return p[i].Value.Tags[metricLogRate] < p[j].Value.Tags[metricLogRate]
	case sortByLogRateLimit:
		return p[i].Value.Tags[metricLogRateLimit] < p[j].Value.Tags[metricLogRateLimit]
	case sortByLogRep:
		return p[i].Value.LogRep < p[j].Value.LogRep
	case sortByLogRtr:
		return p[i].Value.LogRtr < p[j].Value.LogRtr
	case sortByOrg:
		return p[i].Value.OrgName < p[j].Value.OrgName
	case sortBySpace:
		return p[i].Value.SpaceName < p[j].Value.SpaceName
	case sortByIx:
		return p[i].Value.IxCount < p[j].Value.IxCount
	}
	return p[i].Value.Tags[metricAge] > p[j].Value.Tags[metricAge] // default
}
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func passFilter(pairList Pair) bool {
	filterPassed := true
	filterRegex := regexp.MustCompile(common.FilterStrings[filterFieldAppName])
	if !(common.FilterStrings[filterFieldAppName] == "") && !filterRegex.MatchString(pairList.Value.AppName) {
		filterPassed = false
	}
	filterRegex = regexp.MustCompile(common.FilterStrings[filterFieldSpace])
	if !(common.FilterStrings[filterFieldSpace] == "") && !filterRegex.MatchString(pairList.Value.SpaceName) {
		filterPassed = false
	}
	filterRegex = regexp.MustCompile(common.FilterStrings[filterFieldOrg])
	if !(common.FilterStrings[filterFieldOrg] == "") && !filterRegex.MatchString(pairList.Value.OrgName) {
		filterPassed = false
	}
	return filterPassed
}
