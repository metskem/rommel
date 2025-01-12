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
	LastSeen time.Time
	Index    string
	Job      string
	IP       string
	//UpTime                  float64
	//CapacityTotalMemory     float64
	//CapacityAllocatedMemory float64
	//ContainerUsageMemory    float64
	//CapacityTotalDisk       float64
	//ContainerUsageDisk      float64
	//ContainerCount          float64
	//IPTablesRuleCount       float64
	//NetInterfaceCount       float64
	//OverlayTxBytes          float64
	//OverlayRxBytes          float64
	//OverlayTxDropped        float64
	//OverlayRxDropped        float64
	//HTTPRouteCount          float64
	//DopplerConnections      float64
	//ActiveDrains            float64
	//NumCPUS                 float64
	//Responses               float64
	Tags map[string]float64
}

const (
	filterFieldIP int = iota
	TagIP             = "ip"
	TagIx             = "index"
	TagJob            = "job"
)

var (
	mainView      *gocui.View
	summaryView   *gocui.View
	CellMetricMap = make(map[string]CellMetric) // map key is app-guid

	metricAge                     = "container_age"
	metricUpTime                  = "uptime"
	metricContainerUsageMemory    = "ContainerUsageMemory"
	metricCapacityTotalDisk       = "CapacityTotalDisk"
	metricContainerUsageDisk      = "ContainerUsageDisk"
	metricContainerCount          = "ContainerCount"
	metricCapacityTotalMemory     = "CapacityTotalMemory"
	metricCapacityAllocatedMemory = "CapacityAllocatedMemory"
	metricIPTablesRuleCount       = "IPTablesRuleCount"
	metricNetInterfaceCount       = "NetInterfaceCount"
	metricOverlayTxBytes          = "OverlayTxBytes"
	metricOverlayRxBytes          = "OverlayRxBytes"
	metricOverlayRxDropped        = "OverlayRxDropped"
	metricOverlayTxDropped        = "OverlayTxDropped"
	metricHTTPRouteCount          = "HTTPRouteCount"
	metricDopplerConnections      = "doppler_connections"
	metricActiveDrains            = "active_drains"
	metricNumCPUS                 = "numCPUS"
	metricResponses               = "responses"
	metric2xx                     = "responses.2xx"
	metric3xx                     = "responses.3xx"
	metric4xx                     = "responses.4xx"
	metric5xx                     = "responses.5xx"
	MetricNames                   = []string{TagJob, TagIP, metricAge, metricUpTime, metricContainerUsageMemory, metricCapacityTotalDisk, metricContainerUsageDisk, metricContainerCount, metricCapacityTotalMemory, metricIPTablesRuleCount, metricNetInterfaceCount, metricOverlayTxBytes, metricOverlayRxBytes, metricHTTPRouteCount, metricOverlayRxDropped, metricOverlayTxDropped, metricNumCPUS, metricResponses, metric2xx, metric3xx, metric4xx, metric5xx}
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
			if activeSortFieldColor == sortByIP {
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
		defer common.MapLock.Unlock()
		lineCounter := 0
		mainView.Title = "VMs"
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%8s %13s %-14s %13s %8s %7s %10s %6s %7s %7s %7s %5s %5s %5s %5s %6s %8s %8s %6s %6s %6s %6s %6s %s\n", common.ColorYellow,
			"LASTSEEN", "Job", "IP", "UpTime", "NumCPU", "MemTot", "MemAlloc", "MemUsd", "DiskTot", "DiskUsd", "CntrCnt", "IPTR", "NICs", "OVTX", "OVRX", "HTTPRC", "OVRXDrop", "OVTXDrop", "Resp", "2xx", "3xx", "4xx", "5xx", common.ColorReset))
		for _, pairlist := range sortedBy(CellMetricMap, common.ActiveSortDirection, activeSortFieldColor) {
			if passFilter(pairlist) {
				_, _ = fmt.Fprintf(mainView, "%s%8s%s %s%13s%s %s%-14s%s %s%13s%s %s%8s%s %s%7s%s %s%10s%s %s%6s%s %s%7s%s %s%7s%s %s%7s%s %s%5s%s %s%5s%s %s%5s%s %s%5s%s %s%6s%s %s%8s%s %s%8s%s %s%6s%s %s%6s%s %s%6s%s %s%6s%s %s%6s%s\n",
					common.LastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), common.ColorReset,
					JobColor, util.TruncateString(pairlist.Value.Job, 13), common.ColorReset,
					common.IPColor, pairlist.Value.IP, common.ColorReset,
					upTimeColor, util.GetFormattedElapsedTime(1000*1000*1000*pairlist.Value.Tags[metricUpTime]), common.ColorReset,
					numCPUSColor, util.GetFormattedUnit(pairlist.Value.Tags[metricNumCPUS]), common.ColorReset,
					capacityTotalMemoryColor, util.GetFormattedUnit(1024*1024*pairlist.Value.Tags[metricCapacityTotalMemory]), common.ColorReset,
					capacityAllocatedMemoryColor, util.GetFormattedUnit(1024*1024*pairlist.Value.Tags[metricCapacityAllocatedMemory]), common.ColorReset,
					containerUsageMemoryColor, util.GetFormattedUnit(1024*1024*pairlist.Value.Tags[metricContainerUsageMemory]), common.ColorReset,
					CapacityTotalDiskColor, util.GetFormattedUnit(1024*1024*pairlist.Value.Tags[metricCapacityTotalDisk]), common.ColorReset,
					containerUsageDiskColor, util.GetFormattedUnit(1024*1024*pairlist.Value.Tags[metricContainerUsageDisk]), common.ColorReset,
					containerCountColor, util.GetFormattedUnit(pairlist.Value.Tags[metricContainerCount]), common.ColorReset,
					IPTablesRuleCountColor, util.GetFormattedUnit(pairlist.Value.Tags[metricIPTablesRuleCount]), common.ColorReset,
					NetInterfaceCountColor, util.GetFormattedUnit(pairlist.Value.Tags[metricNetInterfaceCount]), common.ColorReset,
					OverlayTxBytesColor, util.GetFormattedUnit(pairlist.Value.Tags[metricOverlayTxBytes]), common.ColorReset,
					OverlayRxBytesColor, util.GetFormattedUnit(pairlist.Value.Tags[metricOverlayRxBytes]), common.ColorReset,
					HTTPRouteCountColor, util.GetFormattedUnit(pairlist.Value.Tags[metricHTTPRouteCount]), common.ColorReset,
					OverlayRxDroppedColor, util.GetFormattedUnit(pairlist.Value.Tags[metricOverlayRxDropped]), common.ColorReset,
					OverlayTxDropped, util.GetFormattedUnit(pairlist.Value.Tags[metricOverlayTxDropped]), common.ColorReset,
					//DopplerConnectionsColor, util.GetFormattedUnit(pairlist.Value.Tags[metricDopplerConnections]), common.ColorReset,
					//ActiveDrainsColor, util.GetFormattedUnit(pairlist.Value.Tags[metricActiveDrains]), common.ColorReset,
					responsesColor, util.GetFormattedUnit(pairlist.Value.Tags[metricResponses]), common.ColorReset,
					r2xxColor, util.GetFormattedUnit(pairlist.Value.Tags[metric2xx]), common.ColorReset,
					r3xxColor, util.GetFormattedUnit(pairlist.Value.Tags[metric3xx]), common.ColorReset,
					r4xxColor, util.GetFormattedUnit(pairlist.Value.Tags[metric4xx]), common.ColorReset,
					r5xxColor, util.GetFormattedUnit(pairlist.Value.Tags[metric5xx]), common.ColorReset,
				)
				lineCounter++
				if lineCounter > maxY-7 {
					//	don't render lines that don't fit on the screen
					break
				}
			}
		}
	}
}
