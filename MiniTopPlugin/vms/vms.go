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
	IP       string
	Tags     map[string]float64
}

const (
	filterFieldIP int = iota
)

var (
	mainView      *gocui.View
	summaryView   *gocui.View
	CellMetricMap map[string]CellMetric // map key is app-guid
	metricIP      = "IP"
	metricAge     = "container_age"
	metricNames   = []string{metricIP}
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
	util.WriteToFile("VMs ShowView")
	colorSortedColumn()
	totalEnvelopesPrev := conf.TotalEnvelopes
	totalEnvelopesRepPrev := conf.TotalEnvelopesRep
	totalEnvelopesRtrPrev := conf.TotalEnvelopesRtr

	// update memory summaries
	var totalMemUsed float64
	conf.MapLock.Lock()
	CellMetricMap = make(map[string]CellMetric)
	for _, metric := range CellMetricMap {
		totalMemUsed += metric.Tags[metricIP]
		updateCellMetrics(&metric)
	}
	conf.MapLock.Unlock()

	gui.Update(func(g *gocui.Gui) error {
		refreshViewContent(g)
		return nil
	})

	conf.TotalEnvelopesPerSec = (conf.TotalEnvelopes - totalEnvelopesPrev) / float64(conf.IntervalSecs)
	conf.TotalEnvelopesRepPerSec = (conf.TotalEnvelopesRep - totalEnvelopesRepPrev) / float64(conf.IntervalSecs)
	conf.TotalEnvelopesRtrPerSec = (conf.TotalEnvelopesRtr - totalEnvelopesRtrPrev) / float64(conf.IntervalSecs)
}

func resetFilters(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	conf.FilterStrings[filterFieldIP] = ""
	return nil
}

func layout(g *gocui.Gui) (err error) {
	if common.ActiveView != common.VMView {
		return nil
	}
	util.WriteToFile("VMs layout")
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
	if conf.ShowFilter {
		if _, err = g.SetView("FilterView", maxX/2-30, maxY/2, maxX/2+30, maxY/2+10, byte(0)); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v, _ := g.SetCurrentView("FilterView")
			v.Title = "Filter"
			_, _ = fmt.Fprint(v, "Filter by (regular expression)")
			if activeSortField == sortByIP {
				_, _ = fmt.Fprintln(v, " IP")
				_, _ = fmt.Fprintln(v, conf.FilterStrings[filterFieldIP])
			}
		}
	}
	if conf.ShowHelp {
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
	util.WriteToFile("VMs refreshViewContent")
	_, maxY := gui.Size()

	summaryView.Clear()
	_, _ = fmt.Fprintf(summaryView, "Target: %s, Nozzle Uptime: %s\n"+
		"Total events: %s (%s/s), RTR events: %s (%s/s), REP events: %s (%s/s), App LogRate: %sBps\n"+
		" Allocated Mem: %s, Used Mem: %s\n",
		conf.ApiAddr, util.GetFormattedElapsedTime((time.Now().Sub(conf.StartTime)).Seconds()*1e9),
		util.GetFormattedUnit(conf.TotalEnvelopes),
		util.GetFormattedUnit(conf.TotalEnvelopesPerSec),
		util.GetFormattedUnit(conf.TotalEnvelopesRtr),
		util.GetFormattedUnit(conf.TotalEnvelopesRtrPerSec),
		util.GetFormattedUnit(conf.TotalEnvelopesRep),
		util.GetFormattedUnit(conf.TotalEnvelopesRepPerSec))

	mainView.Clear()
	conf.MapLock.Lock()
	lineCounter := 0
	mainView.Title = "VMs"
	_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %3s %s\n", conf.ColorYellow, "IP", "LASTSEEN", "IX", conf.ColorReset))
	for _, pairlist := range sortedBy(CellMetricMap, common.ActiveSortDirection, activeSortField) {
		if passFilter(pairlist) {
			_, _ = fmt.Fprintf(mainView, "%s%-20s%s %s%12s%s\n",
				common.LastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), conf.ColorReset,
				common.IPColor, pairlist.Value.IP, conf.ColorReset)
			lineCounter++
			if lineCounter > maxY-7 {
				//	don't render lines that don't fit on the screen
				break
			}
		}
	}
	conf.MapLock.Unlock()
}

// UpdateCellMetrics - Populate the CellMap with the latest cell metrics. */
func updateCellMetrics(cellMetric *CellMetric) {
	cellMetric = &CellMetric{
		LastSeen: cellMetric.LastSeen,
		Index:    cellMetric.Index,
		IP:       cellMetric.IP,
		Tags:     make(map[string]float64),
	}
	for _, metricName := range metricNames {
		cellMetric.Tags[metricName] = cellMetric.Tags[metricName]
	}
}
