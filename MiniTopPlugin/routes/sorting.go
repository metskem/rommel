package routes

import (
	"fmt"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"regexp"
	"sort"
)

const (
	sortByLastSeen = iota
	sortByRoute
)

var (
	activeSortField SortField = sortByRoute
	routeColor                = common.ColorWhite
	r2xxColor                 = common.ColorWhite
)

// based on https://stackoverflow.com/questions/18695346/how-to-sort-a-mapstringint-by-its-values
type SortField int

func sortedBy(metricMap map[string]RouteMetric, reverse bool, sortField SortField) PairList {
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
	Value  RouteMetric
}

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool {
	switch p[i].SortBy {
	case sortByLastSeen:
		return p[i].Value.LastSeen.Unix() < p[j].Value.LastSeen.Unix()
	case sortByRoute:
		return p[i].Value.Route < p[j].Value.Route

	}
	return p[i].Value.Route > p[j].Value.Route // default
}
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func passFilter(pairList Pair) bool {
	filterPassed := true
	filterRegex := regexp.MustCompile(common.FilterStrings[common.FilterFieldIP])
	if !(common.FilterStrings[common.FilterFieldHost] == "") && !filterRegex.MatchString(pairList.Value.Route) {
		filterPassed = false
	}
	return filterPassed
}

func colorSortedColumn() {
	common.LastSeenColor = common.ColorWhite
	routeColor = common.ColorWhite
	switch activeSortField {
	case sortByLastSeen:
		common.LastSeenColor = common.ColorBlue
	case sortByRoute:
		routeColor = common.ColorBlue
	}
	util.WriteToFileDebug(fmt.Sprintf("colorSortedColumn Routes, activeSortField: %d", activeSortField))
}
