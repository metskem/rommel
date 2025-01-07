package apps

import (
	"errors"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"time"
)

var (
	mainView    *gocui.View
	summaryView *gocui.View
)

func SetKeyBindings(gui *gocui.Gui) {
	_ = gui.SetKeybinding("ApplicationView", gocui.KeyArrowRight, gocui.ModNone, arrowRight)
	_ = gui.SetKeybinding("ApplicationView", gocui.KeyArrowLeft, gocui.ModNone, arrowLeft)
	_ = gui.SetKeybinding("ApplicationView", gocui.KeySpace, gocui.ModNone, spacePressed)
	_ = gui.SetKeybinding("ApplicationView", 'f', gocui.ModNone, common.ShowFilterView)
	_ = gui.SetKeybinding("ApplicationView", 't', gocui.ModNone, common.ToggleView)
	_ = gui.SetKeybinding("FilterView", gocui.KeyBackspace, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	_ = gui.SetKeybinding("FilterView", gocui.KeyBackspace2, gocui.ModNone, mkEvtHandler(rune(gocui.KeyBackspace)))
	_ = gui.SetKeybinding("", 'R', gocui.ModNone, resetFilters)
	for _, c := range "\\/[]*?.-@#$%^abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" {
		_ = gui.SetKeybinding("FilterView", c, gocui.ModNone, mkEvtHandler(c))
	}
}

type AppView struct {
}

func NewAppView() *AppView {
	return &AppView{}
}

func (a *AppView) Layout(g *gocui.Gui) error {
	return appLayout(g)
}

func ShowView(gui *gocui.Gui) {
	colorSortedColumn()
	totalEnvelopesPrev := conf.TotalEnvelopes
	totalEnvelopesRepPrev := conf.TotalEnvelopesRep
	totalEnvelopesRtrPrev := conf.TotalEnvelopesRtr

	// update memory summaries
	var totalMemUsed, totalMemAllocated, totalLogRateUsed float64
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

	gui.Update(func(g *gocui.Gui) error {
		refreshViewContent(g)
		return nil
	})

	time.Sleep(time.Duration(conf.IntervalSecs) * time.Second)

	conf.TotalEnvelopesPerSec = (conf.TotalEnvelopes - totalEnvelopesPrev) / float64(conf.IntervalSecs)
	conf.TotalEnvelopesRepPerSec = (conf.TotalEnvelopesRep - totalEnvelopesRepPrev) / float64(conf.IntervalSecs)
	conf.TotalEnvelopesRtrPerSec = (conf.TotalEnvelopesRtr - totalEnvelopesRtrPrev) / float64(conf.IntervalSecs)
}

func resetFilters(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	conf.FilterStrings[conf.FilterFieldAppName] = ""
	conf.FilterStrings[conf.FilterFieldOrg] = ""
	conf.FilterStrings[conf.FilterFieldSpace] = ""
	return nil
}

func appLayout(g *gocui.Gui) (err error) {
	maxX, maxY := g.Size()
	if summaryView, err = g.SetView("SummaryView", 0, 0, maxX-1, 4, byte(0)); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v, _ := g.SetCurrentView("SummaryView")
		v.Title = "Summary"
	}
	if mainView, err = g.SetView("ApplicationView", 0, 5, maxX-1, maxY-1, byte(0)); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v, _ := g.SetCurrentView("ApplicationView")
		v.Title = "Application Instances"
	}
	if conf.ShowFilter {
		if _, err = g.SetView("FilterView", maxX/2-30, maxY/2, maxX/2+30, maxY/2+10, byte(0)); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v, _ := g.SetCurrentView("FilterView")
			v.Title = "Filter"
			_, _ = fmt.Fprint(v, "Filter by (regular expression)")
			if util.ActiveAppsSortField == util.SortByAppName || util.ActiveInstancesSortField == util.SortByAppName {
				_, _ = fmt.Fprintln(v, " AppName")
				_, _ = fmt.Fprintln(v, conf.FilterStrings[conf.FilterFieldAppName])
			}
			if util.ActiveAppsSortField == util.SortBySpace || util.ActiveInstancesSortField == util.SortBySpace {
				_, _ = fmt.Fprintln(v, " Space")
				_, _ = fmt.Fprintln(v, conf.FilterStrings[conf.FilterFieldSpace])
			}
			if util.ActiveAppsSortField == util.SortByOrg || util.ActiveInstancesSortField == util.SortByOrg {
				_, _ = fmt.Fprintln(v, " Org")
				_, _ = fmt.Fprintln(v, conf.FilterStrings[conf.FilterFieldOrg])
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
	_, maxY := gui.Size()

	summaryView.Clear()
	_, _ = fmt.Fprintf(summaryView, "Target: %s, Nozzle Uptime: %s\n"+
		"Total events: %s (%s/s), RTR events: %s (%s/s), REP events: %s (%s/s), App LogRate: %sBps\n"+
		"Total Apps: %d, Instances: %d, Allocated Mem: %s, Used Mem: %s\n",
		conf.ApiAddr, util.GetFormattedElapsedTime((time.Now().Sub(conf.StartTime)).Seconds()*1e9),
		util.GetFormattedUnit(conf.TotalEnvelopes),
		util.GetFormattedUnit(conf.TotalEnvelopesPerSec),
		util.GetFormattedUnit(conf.TotalEnvelopesRtr),
		util.GetFormattedUnit(conf.TotalEnvelopesRtrPerSec),
		util.GetFormattedUnit(conf.TotalEnvelopesRep),
		util.GetFormattedUnit(conf.TotalEnvelopesRepPerSec),
		util.GetFormattedUnit(conf.TotalLogRateUsed/8),
		len(conf.TotalApps),
		len(conf.InstanceMetricMap),
		util.GetFormattedUnit(conf.TotalMemoryAllocated),
		util.GetFormattedUnit(conf.TotalMemoryUsed))

	mainView.Clear()
	conf.MapLock.Lock()
	lineCounter := 0
	if conf.ActiveView == conf.AppInstanceView {
		mainView.Title = "Application Instances"
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %12s %5s %9s %7s %9s %6s %6s %9s %7s %-14s %9s %9s %-25s %-35s%s\n", conf.ColorYellow, "APP/INDEX", "LASTSEEN", "AGE", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "IP", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
		for _, pairlist := range util.SortedBy(conf.InstanceMetricMap, util.ActiveSortDirection, util.ActiveInstancesSortField) {
			if util.PassFilter(pairlist) {
				_, _ = fmt.Fprintf(mainView, "%s%-50s%s %s%5s%s %s%12s%s %s%5s%s %s%9s%s %s%7s%s %s%9s%s %s%6s%s %s%6s%s %s%9s%s %s%7s%s %s%-14s%s %s%9s%s %s%9s%s %s%-25s%s %s%-35s%s\n",
					appNameColor, fmt.Sprintf("%s/%s(%d)", util.TruncateString(pairlist.Value.AppName, 45), pairlist.Value.AppIndex, conf.AppInstanceCounters[pairlist.Value.AppGuid].Count), conf.ColorReset,
					common.LastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), conf.ColorReset,
					common.AgeColor, util.GetFormattedElapsedTime(pairlist.Value.Tags[conf.MetricAge]), conf.ColorReset,
					cpuPercColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricCpu]), conf.ColorReset,
					cpuTotColor, util.GetFormattedUnit(pairlist.Value.CpuTot), conf.ColorReset,
					memoryColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricMemory]), conf.ColorReset,
					memoryLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricMemoryQuota]), conf.ColorReset,
					diskColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricDisk]), conf.ColorReset,
					logRateColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricLogRate]), conf.ColorReset,
					logRateLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricLogRateLimit]), conf.ColorReset,
					entColor, util.GetFormattedUnit(pairlist.Value.Tags[conf.MetricCpuEntitlement]), conf.ColorReset,
					common.IPColor, pairlist.Value.IP, conf.ColorReset,
					logRepColor, util.GetFormattedUnit(pairlist.Value.LogRep), conf.ColorReset,
					logRtrColor, util.GetFormattedUnit(pairlist.Value.LogRtr), conf.ColorReset,
					orgColor, util.TruncateString(pairlist.Value.OrgName, 25), conf.ColorReset,
					spaceColor, pairlist.Value.SpaceName, conf.ColorReset)
				lineCounter++
				if lineCounter > maxY-7 {
					//	don't render lines that don't fit on the screen
					break
				}
			}
		}
	}
	if conf.ActiveView == conf.AppView {
		mainView.Title = "Applications"
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %3s %4s %7s %8s %9s %5s %5s %9s %8s %7s %8s %-25s %-35s%s\n", conf.ColorYellow, "APP", "LASTSEEN", "IX", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
		for _, pairlist := range util.SortedBy(conf.AppMetricMap, util.ActiveSortDirection, util.ActiveAppsSortField) {
			if util.PassFilter(pairlist) {
				_, _ = fmt.Fprintf(mainView, "%s%-50s%s %s%5s%s %s%3d%s %s%4s%s %s%7s%s %s%8s%s %s%9s%s %s%5s%s %s%5s%s %s%9s%s %s%8s%s %s%7s%s %s%8s%s %s%-25s%s %s%-35s%s\n",
					appNameColor, fmt.Sprintf("%s", util.TruncateString(pairlist.Value.AppName, 45)), conf.ColorReset,
					common.LastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), conf.ColorReset,
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
					orgColor, util.TruncateString(pairlist.Value.OrgName, 25), conf.ColorReset,
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

func mkEvtHandler(ch rune) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if util.ActiveInstancesSortField == util.SortByAppName || util.ActiveAppsSortField == util.SortByAppName {
			if ch == rune(gocui.KeyBackspace) {
				if len(conf.FilterStrings[conf.FilterFieldAppName]) > 0 {
					conf.FilterStrings[conf.FilterFieldAppName] = conf.FilterStrings[conf.FilterFieldAppName][:len(conf.FilterStrings[conf.FilterFieldAppName])-1]
					_ = v.SetCursor(len(conf.FilterStrings[conf.FilterFieldAppName])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				conf.FilterStrings[conf.FilterFieldAppName] = conf.FilterStrings[conf.FilterFieldAppName] + string(ch)
			}
		}
		if util.ActiveInstancesSortField == util.SortBySpace || util.ActiveAppsSortField == util.SortBySpace {
			if ch == rune(gocui.KeyBackspace) {
				if len(conf.FilterStrings[conf.FilterFieldSpace]) > 0 {
					conf.FilterStrings[conf.FilterFieldSpace] = conf.FilterStrings[conf.FilterFieldSpace][:len(conf.FilterStrings[conf.FilterFieldSpace])-1]
					_ = v.SetCursor(len(conf.FilterStrings[conf.FilterFieldSpace])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				conf.FilterStrings[conf.FilterFieldSpace] = conf.FilterStrings[conf.FilterFieldSpace] + string(ch)
			}
		}
		if util.ActiveInstancesSortField == util.SortByOrg || util.ActiveAppsSortField == util.SortByOrg {
			if ch == rune(gocui.KeyBackspace) {
				if len(conf.FilterStrings[conf.FilterFieldOrg]) > 0 {
					conf.FilterStrings[conf.FilterFieldOrg] = conf.FilterStrings[conf.FilterFieldOrg][:len(conf.FilterStrings[conf.FilterFieldOrg])-1]
					_ = v.SetCursor(len(conf.FilterStrings[conf.FilterFieldOrg])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				conf.FilterStrings[conf.FilterFieldOrg] = conf.FilterStrings[conf.FilterFieldOrg] + string(ch)
			}
		}
		return nil
	}
}
