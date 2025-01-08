package apps

import (
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"regexp"
	"sort"
)

var (
	appNameColor                       = conf.ColorWhite
	cpuPercColor                       = conf.ColorWhite
	ixColor                            = conf.ColorWhite
	cpuTotColor                        = conf.ColorWhite
	memoryColor                        = conf.ColorWhite
	memoryLimitColor                   = conf.ColorWhite
	diskColor                          = conf.ColorWhite
	logRateColor                       = conf.ColorWhite
	logRateLimitColor                  = conf.ColorWhite
	entColor                           = conf.ColorWhite
	logRepColor                        = conf.ColorWhite
	logRtrColor                        = conf.ColorWhite
	orgColor                           = conf.ColorWhite
	spaceColor                         = conf.ColorWhite
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
	appNameColor = conf.ColorWhite
	common.LastSeenColor = conf.ColorWhite
	common.AgeColor = conf.ColorWhite
	cpuPercColor = conf.ColorWhite
	ixColor = conf.ColorWhite
	cpuTotColor = conf.ColorWhite
	memoryColor = conf.ColorWhite
	memoryLimitColor = conf.ColorWhite
	diskColor = conf.ColorWhite
	logRateColor = conf.ColorWhite
	logRateLimitColor = conf.ColorWhite
	entColor = conf.ColorWhite
	common.IPColor = conf.ColorWhite
	logRepColor = conf.ColorWhite
	logRtrColor = conf.ColorWhite
	orgColor = conf.ColorWhite
	spaceColor = conf.ColorWhite
	if common.ActiveView == common.AppInstanceView {
		switch activeInstancesSortField {
		case sortByAppName:
			appNameColor = conf.ColorBlue
		case sortByLastSeen:
			common.LastSeenColor = conf.ColorBlue
		case sortByAge:
			common.AgeColor = conf.ColorBlue
		case sortByIx:
			ixColor = conf.ColorBlue
		case sortByCpuPerc:
			cpuPercColor = conf.ColorBlue
		case sortByCpuTot:
			cpuTotColor = conf.ColorBlue
		case sortByMemory:
			memoryColor = conf.ColorBlue
		case sortByMemoryLimit:
			memoryLimitColor = conf.ColorBlue
		case sortByDisk:
			diskColor = conf.ColorBlue
		case sortByLogRate:
			logRateColor = conf.ColorBlue
		case sortByLogRateLimit:
			logRateLimitColor = conf.ColorBlue
		case sortByIP:
			common.IPColor = conf.ColorBlue
		case sortByEntitlement:
			entColor = conf.ColorBlue
		case sortByLogRep:
			logRepColor = conf.ColorBlue
		case sortByLogRtr:
			logRtrColor = conf.ColorBlue
		case sortByOrg:
			orgColor = conf.ColorBlue
		case sortBySpace:
			spaceColor = conf.ColorBlue
		}
	}
	if common.ActiveView == common.AppView {
		switch activeAppsSortField {
		case sortByAppName:
			appNameColor = conf.ColorBlue
		case sortByLastSeen:
			common.LastSeenColor = conf.ColorBlue
		case sortByAge:
			common.AgeColor = conf.ColorBlue
		case sortByIx:
			ixColor = conf.ColorBlue
		case sortByCpuPerc:
			cpuPercColor = conf.ColorBlue
		case sortByCpuTot:
			cpuTotColor = conf.ColorBlue
		case sortByMemory:
			memoryColor = conf.ColorBlue
		case sortByMemoryLimit:
			memoryLimitColor = conf.ColorBlue
		case sortByDisk:
			diskColor = conf.ColorBlue
		case sortByLogRate:
			logRateColor = conf.ColorBlue
		case sortByLogRateLimit:
			logRateLimitColor = conf.ColorBlue
		case sortByIP:
			common.IPColor = conf.ColorBlue
		case sortByEntitlement:
			entColor = conf.ColorBlue
		case sortByLogRep:
			logRepColor = conf.ColorBlue
		case sortByLogRtr:
			logRtrColor = conf.ColorBlue
		case sortByOrg:
			orgColor = conf.ColorBlue
		case sortBySpace:
			spaceColor = conf.ColorBlue
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
	filterRegex := regexp.MustCompile(conf.FilterStrings[filterFieldAppName])
	if !(conf.FilterStrings[filterFieldAppName] == "") && !filterRegex.MatchString(pairList.Value.AppName) {
		filterPassed = false
	}
	filterRegex = regexp.MustCompile(conf.FilterStrings[filterFieldSpace])
	if !(conf.FilterStrings[filterFieldSpace] == "") && !filterRegex.MatchString(pairList.Value.SpaceName) {
		filterPassed = false
	}
	filterRegex = regexp.MustCompile(conf.FilterStrings[filterFieldOrg])
	if !(conf.FilterStrings[filterFieldOrg] == "") && !filterRegex.MatchString(pairList.Value.OrgName) {
		filterPassed = false
	}
	return filterPassed
}
