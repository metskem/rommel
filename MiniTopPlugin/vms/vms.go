package vms

import (
	"errors"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"time"
)

type CellMetric struct {
	LastSeen                time.Time
	Index                   string
	IP                      string
	ContainerUsageMemory    float64
	ContainerUsageDisk      float64
	ContainerCount          float64
	CapacityAllocatedMemory float64
	IPTablesRuleCount       float64
	NetInterfaceCount       float64
	Tags                    map[string]float64
}

const (
	filterFieldIP int = iota
	TagIP             = "ip"
	TagIx             = "index"
	TagJob            = "job"
)

var (
	mainView                      *gocui.View
	summaryView                   *gocui.View
	CellMetricMap                 = make(map[string]CellMetric) // map key is app-guid
	metricIP                      = "IP"
	metricAge                     = "container_age"
	MetricContainerUsageMemory    = "ContainerUsageMemory"
	MetricContainerUsageDisk      = "ContainerUsageDisk"
	MetricContainerCount          = "ContainerCount"
	MetricCapacityAllocatedMemory = "CapacityAllocatedMemory"
	MetricIPTablesRuleCount       = "IPTablesRuleCount"
	MetricNetInterfaceCount       = "NetInterfaceCount"
	MetricNames                   = []string{metricIP, metricAge, MetricContainerUsageMemory, MetricContainerUsageDisk, MetricContainerCount, MetricCapacityAllocatedMemory, MetricIPTablesRuleCount, MetricNetInterfaceCount}
)

func SetKeyBindings(gui *gocui.Gui) {
	//_ = gui.SetKeybinding("VMView", gocui.KeyArrowRight, gocui.ModNone, arrowRight)
	//_ = gui.SetKeybinding("VMView", gocui.KeyArrowLeft, gocui.ModNone, arrowLeft)
	_ = gui.SetKeybinding("VMView", gocui.KeySpace, gocui.ModNone, spacePressed)
	_ = gui.SetKeybinding("VMView", 'f', gocui.ModNone, common.ShowFilterView)
	//_ = gui.SetKeybinding("FilterView", gocui.KeyBackspace, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	//_ = gui.SetKeybinding("FilterView", gocui.KeyBackspace2, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	_ = gui.SetKeybinding("", 'R', gocui.ModNone, resetFilters)
	//for _, c := range "\\/[]*?.-@#$%^abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" {
	//_ = gui.SetKeybinding("FilterView", c, gocui.ModNone, mkEvtHandler(c))
	//}
}

type VMView struct {
}

func NewVMView() *VMView {
	return &VMView{}
}

func (a *VMView) Layout(g *gocui.Gui) error {
	return layout(g)
}

func ShowView(gui *gocui.Gui) {
	colorSortedColumn()

	gui.Update(func(g *gocui.Gui) error {
		refreshViewContent(g)
		return nil
	})
}

func resetFilters(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	common.FilterStrings[filterFieldIP] = ""
	return nil
}

func layout(g *gocui.Gui) (err error) {
	if common.ActiveView != common.VMView {
		return nil
	}
	maxX, maxY := g.Size()
	if summaryView, err = g.SetView("SummaryView", 0, 0, maxX-1, 4, byte(0)); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v, _ := g.SetCurrentView("SummaryView")
		v.Title = "Summary"
	}
	if mainView, err = g.SetView("VMView", 0, 5, maxX-1, maxY-1, byte(0)); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v, _ := g.SetCurrentView("VMView")
		v.Title = "VMs"
	}
	if common.ShowFilter {
		if _, err = g.SetView("FilterView", maxX/2-30, maxY/2, maxX/2+30, maxY/2+10, byte(0)); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v, _ := g.SetCurrentView("FilterView")
			v.Title = "Filter"
			_, _ = fmt.Fprint(v, "Filter by (regular expression)")
			if activeSortField == sortByIP {
				_, _ = fmt.Fprintln(v, " IP")
				_, _ = fmt.Fprintln(v, common.FilterStrings[filterFieldIP])
			}
		}
	}
	if common.ShowHelp {
		if _, err = g.SetView("HelpView", maxX/2-40, maxY/2-5, maxX/2+40, maxY/2+15, byte(0)); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v, _ := g.SetCurrentView("HelpView")
			v.Title = "Help"
			_, _ = fmt.Fprintln(v, "You can use the following keys:\nh or ? - show this help (<enter> to close)\nq - quit\nf - filter\nR - reset all filters\narrow keys (left/right) - sort\nspace - flip sort order\nt - toggle between app and instance view")
		}
	}
	return nil
}

func refreshViewContent(gui *gocui.Gui) {
	_, maxY := gui.Size()

	if summaryView != nil {
		summaryView.Clear()
		_, _ = fmt.Fprintf(summaryView, "Target: %s, Nozzle Uptime: %s\n",
			conf.ApiAddr, util.GetFormattedElapsedTime((time.Now().Sub(common.StartTime)).Seconds()*1e9))
	}
	if mainView != nil {
		mainView.Clear()
		common.MapLock.Lock()
		lineCounter := 0
		mainView.Title = "VMs"
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%8s %-14s %8s %9s %10s %9s %13s %12s %s\n", common.ColorYellow,
			"LASTSEEN", "IP", "AllocMem", "CntrMemUse", "CntrDiskUse", "CntrCnt", "IPTablesRules", "NetIntrfcCnt", common.ColorReset))
		for _, pairlist := range sortedBy(CellMetricMap, common.ActiveSortDirection, activeSortField) {
			if passFilter(pairlist) {
				_, _ = fmt.Fprintf(mainView, "%s%8s%s %s%-14s%s %s%8s%s %s%10s%s %s%11s%s %s%9s%s %s%13s%s %s%12s%s\n",
					common.LastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), common.ColorReset,
					common.IPColor, pairlist.Value.IP, common.ColorReset,
					capacityAllocatedMemoryColor, util.GetFormattedUnit(pairlist.Value.Tags[MetricCapacityAllocatedMemory]), common.ColorReset,
					containerUsageMemoryColor, util.GetFormattedUnit(pairlist.Value.Tags[MetricContainerUsageMemory]), common.ColorReset,
					containerUsageDiskColor, util.GetFormattedUnit(pairlist.Value.Tags[MetricContainerUsageDisk]), common.ColorReset,
					containerCountColor, util.GetFormattedUnit(pairlist.Value.Tags[MetricContainerCount]), common.ColorReset,
					IPTablesRuleCount, util.GetFormattedUnit(pairlist.Value.Tags[MetricIPTablesRuleCount]), common.ColorReset,
					NetInterfaceCount, util.GetFormattedUnit(pairlist.Value.Tags[MetricNetInterfaceCount]), common.ColorReset,
				)
				lineCounter++
				if lineCounter > maxY-7 {
					//	don't render lines that don't fit on the screen
					break
				}
			}
		}
		common.MapLock.Unlock()
	}
}
