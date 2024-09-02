package cui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/metskem/rommel/StatsNozzleV2/conf"
	"github.com/metskem/rommel/StatsNozzleV2/util"
	"log"
	"os"
	"strings"
	"time"
)

var (
	mainView    *gocui.View
	summaryView *gocui.View
	//filterView        *gocui.View
	g                 *gocui.Gui
	appNameColor      = conf.ColorWhite
	lastSeenColor     = conf.ColorWhite
	ageColor          = conf.ColorWhite
	cpuPercColor      = conf.ColorWhite
	cpuTotColor       = conf.ColorWhite
	memoryColor       = conf.ColorWhite
	memoryLimitColor  = conf.ColorWhite
	diskColor         = conf.ColorWhite
	logRateColor      = conf.ColorWhite
	logRateLimitColor = conf.ColorWhite
	entColor          = conf.ColorWhite
	IPColor           = conf.ColorWhite
	logRepColor       = conf.ColorWhite
	logRtrColor       = conf.ColorWhite
	orgColor          = conf.ColorWhite
	spaceColor        = conf.ColorWhite
)

func Start() {
	var err error
	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	//g.InputEsc = true

	//_ = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	_ = g.SetKeybinding("ApplicationView", 'q', gocui.ModNone, quit)
	_ = g.SetKeybinding("ApplicationView", 'd', gocui.ModNone, dumper)
	_ = g.SetKeybinding("ApplicationView", gocui.KeyArrowRight, gocui.ModNone, arrowRight)
	_ = g.SetKeybinding("ApplicationView", gocui.KeyArrowLeft, gocui.ModNone, arrowLeft)
	_ = g.SetKeybinding("ApplicationView", gocui.KeySpace, gocui.ModNone, spacePressed)
	_ = g.SetKeybinding("ApplicationView", 'f', gocui.ModNone, showFilterView)

	for _, c := range "-@#$%^abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" {
		g.SetKeybinding("FilterView", c, gocui.ModNone, mkEvtHandler(c))
	}

	_ = g.SetKeybinding("FilterView", gocui.KeyBackspace, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	_ = g.SetKeybinding("FilterView", gocui.KeyBackspace2, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	_ = g.SetKeybinding("FilterView", gocui.KeyEnter, gocui.ModNone, handleFilterEnter)
	_ = g.SetKeybinding("FilterView", gocui.KeyEsc, gocui.ModNone, handleFilterEsc)

	//_ = g.SetKeybinding("ApplicationView", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return scrollView(v, -1) })
	//_ = g.SetKeybinding("ApplicationView", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return scrollView(v, 1) })

	colorSortedColumn()

	go func() {
		for {
			g.Update(func(g *gocui.Gui) error {
				refreshViewContent()
				return nil
			})
			time.Sleep(1 * time.Second)
		}
	}()

	if err = g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) (err error) {
	maxX, maxY := g.Size()
	if summaryView, err = g.SetView("SummaryView", 0, 0, maxX-1, 4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v, _ := g.SetCurrentView("SummaryView")
		v.Title = "Summary"
	}
	if mainView, err = g.SetView("ApplicationView", 0, 5, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v, _ := g.SetCurrentView("ApplicationView")
		v.Title = "Application Instances"
	}
	if conf.ShowFilter {
		if _, err = g.SetView("FilterView", maxX/2-30, maxY/2, maxX/2+30, maxY/2+10); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v, _ := g.SetCurrentView("FilterView")
			v.Title = "Filter"
			_, _ = fmt.Fprintln(v, "Filter by:")
			_, _ = fmt.Fprint(v, conf.FilterString)
			//v.Editable = true
			//v.Overwrite = true
		}
	}
	return nil
}

func refreshViewContent() {
	_, maxY := g.Size()

	summaryView.Clear()
	_, _ = fmt.Fprintf(summaryView, "Target: %s, Nozzle Uptime: %s\nTotal events: %s, Total RTR events: %s, Total REP events: %s\nTotal Apps: %d, Total App Instances: %d", conf.ApiAddr, util.GetFormattedElapsedTime((time.Now().Sub(conf.StartTime)).Seconds()*1e9), util.GetFormattedUnit(conf.TotalEnvelopes), util.GetFormattedUnit(conf.TotalEnvelopesRtr), util.GetFormattedUnit(conf.TotalEnvelopesRep), len(conf.TotalApps), len(conf.MetricMap))

	mainView.Clear()
	_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-62s %8s %12s %10s %12s %7s %9s %8s %7s %9s %9s %14s %9s %9s %-25s %-35s%s\n", conf.ColorYellow, "APP/INDEX", "LASTSEEN", "AGE", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "IP", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
	conf.MapLock.Lock()

	lineCounter := 0
	for _, pairlist := range util.SortedBy(conf.MetricMap, util.ActiveSortDirection, util.ActiveSortField) {
		if conf.FilterString == "" || strings.HasPrefix(pairlist.Value.AppName, conf.FilterString) {
			_, _ = fmt.Fprintf(mainView, "%s%-65s%s %s%5s%s %s%12s%s %s%10s%s %s%12s%s %s%7s%s %s%9s%s %s%8s%s %s%7s%s %s%9s%s %s%9s%s %s%14s%s %s%9s%s %s%9s%s %s%-25s%s %s%-35s%s\n",
				appNameColor, fmt.Sprintf("%s/%s(%d)", pairlist.Value.AppName, pairlist.Value.AppIndex, conf.AppInstanceCount[pairlist.Value.AppGuid]), conf.ColorReset,
				lastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), conf.ColorReset,
				ageColor, util.GetFormattedElapsedTime(pairlist.Value.Tags[conf.MetricAge]), conf.ColorReset,
				cpuPercColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricCpu]), conf.ColorReset,
				cpuTotColor, util.GetFormattedUnit(pairlist.Value.CpuTot), conf.ColorReset,
				memoryColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricMemory]), conf.ColorReset,
				memoryLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricMemoryQuota]), conf.ColorReset,
				diskColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricDisk]), conf.ColorReset,
				logRateColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricLogRate]), conf.ColorReset,
				logRateLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricLogRateLimit]), conf.ColorReset,
				entColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricCpuEntitlement]), conf.ColorReset,
				IPColor, pairlist.Value.IP, conf.ColorReset,
				logRepColor, util.GetFormattedUnit(pairlist.Value.LogRep), conf.ColorReset,
				logRtrColor, util.GetFormattedUnit(pairlist.Value.LogRtr), conf.ColorReset,
				orgColor, pairlist.Value.OrgName, conf.ColorReset,
				spaceColor, pairlist.Value.SpaceName, conf.ColorReset)
			lineCounter++
			if lineCounter > maxY-7 {
				//	don't render lines that don't fit on the screen
				break
			}
		}
	}

	conf.MapLock.Unlock()
}

func quit(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	os.Exit(0)
	return gocui.ErrQuit
}
func dumper(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	conf.MapLock.Lock()
	for key, metric := range conf.MetricMap {
		util.WriteToFile(fmt.Sprintf("%s %s %s %s", key, metric.OrgName, metric.SpaceName, metric.AppName))
	}
	conf.MapLock.Unlock()
	return nil
}
func arrowRight(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	if util.ActiveSortField != util.SortBySpace {
		util.ActiveSortField++
	}
	colorSortedColumn()
	return nil
}
func arrowLeft(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	if util.ActiveSortField != util.SortByAppName {
		util.ActiveSortField--
	}
	colorSortedColumn()
	return nil
}
func spacePressed(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	flipSortOrder()
	return nil
}
func showFilterView(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	conf.ShowFilter = true
	return nil
}

func flipSortOrder() {
	if util.ActiveSortDirection == true {
		util.ActiveSortDirection = false
	} else {
		util.ActiveSortDirection = true
	}
}

func colorSortedColumn() {
	appNameColor = conf.ColorWhite
	lastSeenColor = conf.ColorWhite
	ageColor = conf.ColorWhite
	cpuPercColor = conf.ColorWhite
	cpuTotColor = conf.ColorWhite
	memoryColor = conf.ColorWhite
	memoryLimitColor = conf.ColorWhite
	diskColor = conf.ColorWhite
	logRateColor = conf.ColorWhite
	logRateLimitColor = conf.ColorWhite
	entColor = conf.ColorWhite
	IPColor = conf.ColorWhite
	logRepColor = conf.ColorWhite
	logRateColor = conf.ColorWhite
	orgColor = conf.ColorWhite
	spaceColor = conf.ColorWhite
	switch util.ActiveSortField {
	case util.SortByAppName:
		appNameColor = conf.ColorBlue
	case util.SortByLastSeen:
		lastSeenColor = conf.ColorBlue
	case util.SortByAge:
		ageColor = conf.ColorBlue
	case util.SortByCpuPerc:
		cpuPercColor = conf.ColorBlue
	case util.SortByCpuTot:
		cpuTotColor = conf.ColorBlue
	case util.SortByMemory:
		memoryColor = conf.ColorBlue
	case util.SortByMemoryLimit:
		memoryLimitColor = conf.ColorBlue
	case util.SortByDisk:
		diskColor = conf.ColorBlue
	case util.SortByLogRate:
		logRateColor = conf.ColorBlue
	case util.SortByLogRateLimit:
		logRateLimitColor = conf.ColorBlue
	case util.SortByIP:
		IPColor = conf.ColorBlue
	case util.SortByEntitlement:
		entColor = conf.ColorBlue
	case util.SortByLogRep:
		logRepColor = conf.ColorBlue
	case util.SortByLogRtr:
		logRtrColor = conf.ColorBlue
	case util.SortByOrg:
		orgColor = conf.ColorBlue
	case util.SortBySpace:
		spaceColor = conf.ColorBlue
	}
}

func mkEvtHandler(ch rune) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if ch == rune(gocui.KeyBackspace) {
			if len(conf.FilterString) > 0 {
				conf.FilterString = conf.FilterString[:len(conf.FilterString)-1]
				v.SetCursor(len(conf.FilterString)+1, 1)
				v.EditDelete(true)
			}
			return nil
		} else {
			_, _ = fmt.Fprint(v, string(ch))
			conf.FilterString = conf.FilterString + string(ch)
		}
		return nil
	}
}

func handleFilterEnter(g *gocui.Gui, v *gocui.View) error {
	conf.ShowFilter = false
	g.DeleteView("FilterView")
	_, _ = g.SetCurrentView("ApplicationView")
	return nil
}

func handleFilterEsc(g *gocui.Gui, v *gocui.View) error {
	conf.ShowFilter = false
	g.DeleteView("FilterView")
	_, _ = g.SetCurrentView("ApplicationView")
	return nil
}
