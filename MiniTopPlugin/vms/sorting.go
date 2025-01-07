package vms

import (
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"regexp"
	"sort"
)

var (
	ixColor                   = conf.ColorWhite
	activeSortField SortField = sortByIP
)

func spacePressed(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	common.FlipSortOrder()
	return nil
}

func colorSortedColumn() {
	common.LastSeenColor = conf.ColorWhite
	common.AgeColor = conf.ColorWhite
	ixColor = conf.ColorWhite
	common.IPColor = conf.ColorWhite
	switch activeSortField {
	case sortByLastSeen:
		common.LastSeenColor = conf.ColorBlue
	case sortByAge:
		common.AgeColor = conf.ColorBlue
	case sortByIx:
		ixColor = conf.ColorBlue
	case sortByIP:
		common.IPColor = conf.ColorBlue
	}
}

// based on https://stackoverflow.com/questions/18695346/how-to-sort-a-mapstringint-by-its-values
type SortField int

const (
	sortByLastSeen = iota
	sortByAge
	sortByIx
	sortByIP
)

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
	case sortByIP:
		return p[i].Value.IP < p[j].Value.IP
	}
	return p[i].Value.Tags[metricAge] > p[j].Value.Tags[metricAge] // default
}
func (p PairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func passFilter(pairList Pair) bool {
	filterPassed := true
	filterRegex := regexp.MustCompile(conf.FilterStrings[filterFieldIP])
	if !(conf.FilterStrings[filterFieldIP] == "") && !filterRegex.MatchString(pairList.Value.IP) {
		filterPassed = false
	}
	return filterPassed
}
