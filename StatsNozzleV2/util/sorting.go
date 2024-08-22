package util

import (
	"github.com/metskem/rommel/StatsNozzleV2/conf"
	"sort"
)

// based on https://stackoverflow.com/questions/18695346/how-to-sort-a-mapstringint-by-its-values
type SortField int

const (
	SortByAppName = iota
	SortByAge
	SortByCpu
	SortByMemory
	SortByDisk
	SortByLogRate
	SortByEntitlement
	SortByOrg
	SortBySpace
)

var (
	ActiveSortField     SortField = SortByCpu
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
	case SortByAge:
		return p[i].Value.Values[conf.MetricAge] < p[j].Value.Values[conf.MetricAge]
	case SortByCpu:
		return p[i].Value.Values[conf.MetricCpu] < p[j].Value.Values[conf.MetricCpu]
	case SortByMemory:
		return p[i].Value.Values[conf.MetricMemory] < p[j].Value.Values[conf.MetricMemory]
	case SortByDisk:
		return p[i].Value.Values[conf.MetricDisk] < p[j].Value.Values[conf.MetricDisk]
	case SortByEntitlement:
		return p[i].Value.Values[conf.MetricCpuEntitlement] < p[j].Value.Values[conf.MetricCpuEntitlement]
	case SortByLogRate:
		return p[i].Value.Values[conf.MetricLogRate] < p[j].Value.Values[conf.MetricLogRate]
	case SortByOrg:
		return p[i].Value.OrgName < p[j].Value.OrgName
	case SortBySpace:
		return p[i].Value.SpaceName < p[j].Value.SpaceName
	}
	return p[i].Value.Values[conf.MetricAge] > p[j].Value.Values[conf.MetricAge] // default
}
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
