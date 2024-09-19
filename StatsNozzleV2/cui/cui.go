package cui

import (
	"errors"
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/metskem/rommel/StatsNozzleV2/conf"
	"github.com/metskem/rommel/StatsNozzleV2/util"
	"log"
	"os"
	"regexp"
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
	ixColor           = conf.ColorWhite
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
	_ = g.SetKeybinding("ApplicationView", 't', gocui.ModNone, toggleAppOrInstanceView)

	for _, c := range "\\/[]*?.-@#$%^abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" {
		_ = g.SetKeybinding("FilterView", c, gocui.ModNone, mkEvtHandler(c))
	}

	_ = g.SetKeybinding("FilterView", gocui.KeyBackspace, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	_ = g.SetKeybinding("FilterView", gocui.KeyBackspace2, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	_ = g.SetKeybinding("FilterView", gocui.KeyEnter, gocui.ModNone, handleFilterEnter)
	_ = g.SetKeybinding("FilterView", gocui.KeyEsc, gocui.ModNone, handleFilterEsc)

	//_ = g.SetKeybinding("ApplicationView", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return scrollView(v, -1) })
	//_ = g.SetKeybinding("ApplicationView", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error { return scrollView(v, 1) })

	colorSortedColumn()

	//  main UI refresh loop
	go func() {
		for {
			totalEnvelopesPrev := conf.TotalEnvelopes
			totalEnvelopesRepPrev := conf.TotalEnvelopesRep
			totalEnvelopesRtrPrev := conf.TotalEnvelopesRtr
			totalLogRateUsed := conf.TotalLogRateUsed

			// update memory summaries
			var totalMemUsed float64
			var totalMemAllocated float64
			conf.MapLock.Lock()
			conf.AppMetricMap = make(map[string]conf.AppOrInstanceMetric)
			for _, metric := range conf.InstanceMetricMap {
				totalMemUsed += metric.Tags[conf.MetricMemory]
				totalMemAllocated += metric.Tags[conf.MetricMemoryQuota]
				totalLogRateUsed += metric.Tags[conf.MetricLogRate]
				util.UpdateAppMetrics(&metric)
			}
			conf.MapLock.Unlock()
			conf.TotalMemoryUsed = totalMemUsed
			conf.TotalMemoryAllocated = totalMemAllocated
			conf.TotalLogRateUsed = totalLogRateUsed

			g.Update(func(g *gocui.Gui) error {
				refreshViewContent()
				return nil
			})

			time.Sleep(time.Duration(conf.IntervalSecs) * time.Second)

			conf.TotalEnvelopesPerSec = (conf.TotalEnvelopes - totalEnvelopesPrev) / float64(conf.IntervalSecs)
			conf.TotalEnvelopesRepPerSec = (conf.TotalEnvelopesRep - totalEnvelopesRepPrev) / float64(conf.IntervalSecs)
			conf.TotalEnvelopesRtrPerSec = (conf.TotalEnvelopesRtr - totalEnvelopesRtrPrev) / float64(conf.IntervalSecs)

		}
	}()

	if err = g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) (err error) {
	maxX, maxY := g.Size()
	if summaryView, err = g.SetView("SummaryView", 0, 0, maxX-1, 4); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v, _ := g.SetCurrentView("SummaryView")
		v.Title = "Summary"
	}
	if mainView, err = g.SetView("ApplicationView", 0, 5, maxX-1, maxY-1); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v, _ := g.SetCurrentView("ApplicationView")
		v.Title = "Application Instances"
	}
	if conf.ShowFilter {
		if _, err = g.SetView("FilterView", maxX/2-30, maxY/2, maxX/2+30, maxY/2+10); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v, _ := g.SetCurrentView("FilterView")
			v.Title = "Filter"
			_, _ = fmt.Fprintln(v, "Filter by (regular expression):")
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
	_, _ = fmt.Fprintf(summaryView, "Target: %s, Nozzle Uptime: %s\n"+
		"Total events: %s (%s/s), RTR events: %s (%s/s), REP events: %s (%s/s), App LogRate: %sBps\n"+
		"Total Apps: %d, Instances: %d, Allocated Mem: %s, Used Mem: %s\n",
		conf.ApiAddr, util.GetFormattedElapsedTime((time.Now().Sub(conf.StartTime)).Seconds()*1e9), util.GetFormattedUnit(conf.TotalEnvelopes), util.GetFormattedUnit(conf.TotalEnvelopesPerSec), util.GetFormattedUnit(conf.TotalEnvelopesRtr), util.GetFormattedUnit(conf.TotalEnvelopesRtrPerSec), util.GetFormattedUnit(conf.TotalEnvelopesRep), util.GetFormattedUnit(conf.TotalEnvelopesRepPerSec), util.GetFormattedUnit(conf.TotalLogRateUsed/8), len(conf.TotalApps), len(conf.InstanceMetricMap), util.GetFormattedUnit(conf.TotalMemoryAllocated), util.GetFormattedUnit(conf.TotalMemoryUsed))

	mainView.Clear()
	conf.MapLock.Lock()
	lineCounter := 0
	filterRegex := regexp.MustCompile(conf.FilterString)
	if conf.AppOrInstanceView == conf.AppOrInstanceViewInstance {
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %12s %5s %9s %7s %9s %6s %6s %9s %7s %-14s %9s %9s %-25s %-35s%s\n", conf.ColorYellow, "APP/INDEX", "LASTSEEN", "AGE", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "IP", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
		for _, pairlist := range util.SortedBy(conf.InstanceMetricMap, util.ActiveSortDirection, util.ActiveSortField) {
			if conf.FilterString == "" || filterRegex.MatchString(pairlist.Value.AppName) {
				_, _ = fmt.Fprintf(mainView, "%s%-50s%s %s%5s%s %s%12s%s %s%5s%s %s%9s%s %s%7s%s %s%9s%s %s%6s%s %s%6s%s %s%9s%s %s%7s%s %s%-14s%s %s%9s%s %s%9s%s %s%-25s%s %s%-35s%s\n",
					appNameColor, fmt.Sprintf("%s/%s(%d)", pairlist.Value.AppName, pairlist.Value.AppIndex, conf.AppInstanceCounters[pairlist.Value.AppGuid].Count), conf.ColorReset,
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
	}
	if conf.AppOrInstanceView == conf.AppOrInstanceViewApp {
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %3s %4s %7s %8s %9s %5s %5s %9s %8s %7s %8s %-25s %-35s%s\n", conf.ColorYellow, "APP", "LASTSEEN", "IX", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
		for _, pairlist := range util.SortedBy(conf.AppMetricMap, util.ActiveSortDirection, util.ActiveSortField) {
			if conf.FilterString == "" || filterRegex.MatchString(pairlist.Value.AppName) {
				_, _ = fmt.Fprintf(mainView, "%s%-50s%s %s%5s%s %s%3d%s %s%4s%s %s%7s%s %s%8s%s %s%9s%s %s%5s%s %s%5s%s %s%9s%s %s%8s%s %s%7s%s %s%8s%s %s%-25s%s %s%-35s%s\n",
					appNameColor, fmt.Sprintf("%s", pairlist.Value.AppName), conf.ColorReset,
					lastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), conf.ColorReset,
					ixColor, pairlist.Value.IxCount, conf.ColorReset,
					cpuPercColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricCpu]), conf.ColorReset,
					cpuTotColor, util.GetFormattedUnit(pairlist.Value.CpuTot), conf.ColorReset,
					memoryColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricMemory]), conf.ColorReset,
					memoryLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricMemoryQuota]), conf.ColorReset,
					diskColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricDisk]), conf.ColorReset,
					logRateColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricLogRate]), conf.ColorReset,
					logRateLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricLogRateLimit]), conf.ColorReset,
					entColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricCpuEntitlement]), conf.ColorReset,
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
	for key, metric := range conf.InstanceMetricMap {
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
	ixColor = conf.ColorWhite
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
	case util.SortByIx:
		ixColor = conf.ColorBlue
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
				_ = v.SetCursor(len(conf.FilterString)+1, 1)
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
	_ = v // get rid of compiler warning
	conf.ShowFilter = false
	_ = g.DeleteView("FilterView")
	_, _ = g.SetCurrentView("ApplicationView")
	return nil
}

func toggleAppOrInstanceView(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	util.WriteToFile("toggleAppOrInstanceView")
	if conf.AppOrInstanceView == conf.AppOrInstanceViewInstance {
		conf.AppOrInstanceView = conf.AppOrInstanceViewApp
	} else {
		conf.AppOrInstanceView = conf.AppOrInstanceViewInstance
	}
	return nil
}

func handleFilterEsc(g *gocui.Gui, v *gocui.View) error {
	_ = v // get rid of compiler warning
	conf.ShowFilter = false
	_ = g.DeleteView("FilterView")
	_, _ = g.SetCurrentView("ApplicationView")
	return nil
}
