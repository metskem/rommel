package util

import (
	"github.com/metskem/rommel/StatsNozzleV2/conf"
	"sort"
)

// based on https://stackoverflow.com/questions/18695346/how-to-sort-a-mapstringint-by-its-values
type SortField int

const (
	SortByAppName = iota
	SortByLastSeen
	SortByAge
	SortByCpuPerc
	SortByCpuTot
	SortByMemory
	SortByMemoryLimit
	SortByDisk
	SortByLogRate
	SortByLogRateLimit
	SortByEntitlement
	SortByIP
	SortByLogRep
	SortByLogRtr
	SortByOrg
	SortBySpace
)

var (
	ActiveSortField     SortField = SortByCpuPerc
	ActiveSortDirection           = true
)

func SortedBy(metricMap map[string]conf.Metric, reverse bool, sortField SortField) PairList {
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
	Value  conf.Metric
}

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool {
	switch p[i].SortBy {
	case SortByAppName:
		return p[i].Value.AppName < p[j].Value.AppName
	case SortByLastSeen:
		return p[i].Value.LastSeen.Unix() < p[j].Value.LastSeen.Unix()
	case SortByAge:
		return p[i].Value.Tags[conf.MetricAge] < p[j].Value.Tags[conf.MetricAge]
	case SortByCpuPerc:
		return p[i].Value.Tags[conf.MetricCpu] < p[j].Value.Tags[conf.MetricCpu]
	case SortByCpuTot:
		return p[i].Value.CpuTot < p[j].Value.CpuTot
	case SortByMemory:
		return p[i].Value.Tags[conf.MetricMemory] < p[j].Value.Tags[conf.MetricMemory]
	case SortByMemoryLimit:
		return p[i].Value.Tags[conf.MetricMemoryQuota] < p[j].Value.Tags[conf.MetricMemoryQuota]
	case SortByDisk:
		return p[i].Value.Tags[conf.MetricDisk] < p[j].Value.Tags[conf.MetricDisk]
	case SortByEntitlement:
		return p[i].Value.Tags[conf.MetricCpuEntitlement] < p[j].Value.Tags[conf.MetricCpuEntitlement]
	case SortByIP:
		return p[i].Value.IP < p[j].Value.IP
	case SortByLogRate:
		return p[i].Value.Tags[conf.MetricLogRate] < p[j].Value.Tags[conf.MetricLogRate]
	case SortByLogRateLimit:
		return p[i].Value.Tags[conf.MetricLogRateLimit] < p[j].Value.Tags[conf.MetricLogRateLimit]
	case SortByLogRep:
		return p[i].Value.LogRep < p[j].Value.LogRep
	case SortByLogRtr:
		return p[i].Value.LogRtr < p[j].Value.LogRtr
	case SortByOrg:
		return p[i].Value.OrgName < p[j].Value.OrgName
	case SortBySpace:
		return p[i].Value.SpaceName < p[j].Value.SpaceName
	}
	return p[i].Value.Tags[conf.MetricAge] > p[j].Value.Tags[conf.MetricAge] // default
}
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
