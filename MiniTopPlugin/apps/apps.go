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

type AppOrInstanceMetric struct {
	LastSeen  time.Time
	AppIndex  string
	IxCount   int
	AppName   string
	AppGuid   string
	SpaceName string
	OrgName   string
	CpuTot    float64
	LogRtr    float64
	LogRep    float64
	IP        string
	Tags      map[string]float64
}

type AppInstanceCounter struct {
	Count       int
	LastUpdated time.Time
}

const (
	filterFieldAppName int = iota
	filterFieldOrg
	filterFieldSpace
	MetricCpu            = "cpu"
	metricAge            = "container_age"
	metricCpuEntitlement = "cpu_entitlement"
	metricDisk           = "disk"
	metricMemory         = "memory"
	metricMemoryQuota    = "memory_quota"
	metricLogRate        = "log_rate"
	metricLogRateLimit   = "log_rate_limit"
	TagAppId             = "app_id"
	TagOrgName           = "organization_name"
	TagSpaceName         = "space_name"
	TagAppName           = "app_name"
	TagAppInstanceId     = "instance_id" // use this for app index
	TagOrigin            = "origin"
	TagOriginValueRep    = "rep"
	TagOriginValueRtr    = "gorouter"
)

var (
	mainView             *gocui.View
	summaryView          *gocui.View
	AppMetricMap         map[string]AppOrInstanceMetric         // map key is app-guid
	InstanceMetricMap    = make(map[string]AppOrInstanceMetric) // map key is app-guid/index
	AppInstanceCounters  = make(map[string]AppInstanceCounter)  // here we keep the highest instance index for each app
	TotalApps            = make(map[string]bool)
	totalMemoryUsed      float64
	totalMemoryAllocated float64
	totalLogRateUsed     float64

	MetricNames = []string{MetricCpu, metricAge, metricCpuEntitlement, metricDisk, metricMemory, metricMemoryQuota, metricLogRate, metricLogRateLimit}
)

func SetKeyBindings(gui *gocui.Gui) {
	_ = gui.SetKeybinding("ApplicationView", gocui.KeyArrowRight, gocui.ModNone, arrowRight)
	_ = gui.SetKeybinding("ApplicationView", gocui.KeyArrowLeft, gocui.ModNone, arrowLeft)
	_ = gui.SetKeybinding("", gocui.KeySpace, gocui.ModNone, common.SpacePressed)
	_ = gui.SetKeybinding("", 'f', gocui.ModNone, common.ShowFilterView)
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
	return layout(g)
}

func ShowView(gui *gocui.Gui) {
	colorSortedColumn()
	totalEnvelopesPrev := common.TotalEnvelopes
	totalEnvelopesRepPrev := common.TotalEnvelopesRep
	totalEnvelopesRtrPrev := common.TotalEnvelopesRtr

	// update memory summaries
	var totalMemUsed, totalMemAllocated, totalLogRtUsed float64
	common.MapLock.Lock()
	AppMetricMap = make(map[string]AppOrInstanceMetric)
	for _, metric := range InstanceMetricMap {
		totalMemUsed += metric.Tags[metricMemory]
		totalMemAllocated += metric.Tags[metricMemoryQuota]
		totalLogRtUsed += metric.Tags[metricLogRate]
		updateAppMetrics(&metric)
	}
	common.MapLock.Unlock()
	totalMemoryUsed = totalMemUsed
	totalMemoryAllocated = totalMemAllocated
	totalLogRateUsed = totalLogRtUsed

	gui.Update(func(g *gocui.Gui) error {
		refreshViewContent(g)
		return nil
	})

	common.TotalEnvelopesPerSec = (common.TotalEnvelopes - totalEnvelopesPrev) / float64(conf.IntervalSecs)
	common.TotalEnvelopesRepPerSec = (common.TotalEnvelopesRep - totalEnvelopesRepPrev) / float64(conf.IntervalSecs)
	common.TotalEnvelopesRtrPerSec = (common.TotalEnvelopesRtr - totalEnvelopesRtrPrev) / float64(conf.IntervalSecs)
}

func resetFilters(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	common.FilterStrings[filterFieldAppName] = ""
	common.FilterStrings[filterFieldOrg] = ""
	common.FilterStrings[filterFieldSpace] = ""
	return nil
}

func layout(g *gocui.Gui) (err error) {
	if common.ActiveView != common.AppView && common.ActiveView != common.AppInstanceView {
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
	if mainView, err = g.SetView("ApplicationView", 0, 5, maxX-1, maxY-1, byte(0)); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v, _ := g.SetCurrentView("ApplicationView")
		v.Title = "Application Instances"
	}
	if common.ShowFilter {
		if _, err = g.SetView("FilterView", maxX/2-30, maxY/2, maxX/2+30, maxY/2+10, byte(0)); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v, _ := g.SetCurrentView("FilterView")
			v.Title = "Filter"
			_, _ = fmt.Fprint(v, "Filter by (regular expression)")
			if activeAppsSortField == sortByAppName || activeInstancesSortField == sortByAppName {
				_, _ = fmt.Fprintln(v, " AppName")
				_, _ = fmt.Fprintln(v, common.FilterStrings[filterFieldAppName])
			}
			if activeAppsSortField == sortBySpace || activeInstancesSortField == sortBySpace {
				_, _ = fmt.Fprintln(v, " Space")
				_, _ = fmt.Fprintln(v, common.FilterStrings[filterFieldSpace])
			}
			if activeAppsSortField == sortByOrg || activeInstancesSortField == sortByOrg {
				_, _ = fmt.Fprintln(v, " Org")
				_, _ = fmt.Fprintln(v, common.FilterStrings[filterFieldOrg])
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
		_, _ = fmt.Fprintf(summaryView, "Target: %s, Nozzle Uptime: %s\n"+
			"Total events: %s (%s/s), RTR events: %s (%s/s), REP events: %s (%s/s), App LogRate: %sBps\n"+
			"Total Apps: %d, Instances: %d, Allocated Mem: %s, Used Mem: %s\n",
			conf.ApiAddr, util.GetFormattedElapsedTime((time.Now().Sub(common.StartTime)).Seconds()*1e9),
			util.GetFormattedUnit(common.TotalEnvelopes),
			util.GetFormattedUnit(common.TotalEnvelopesPerSec),
			util.GetFormattedUnit(common.TotalEnvelopesRtr),
			util.GetFormattedUnit(common.TotalEnvelopesRtrPerSec),
			util.GetFormattedUnit(common.TotalEnvelopesRep),
			util.GetFormattedUnit(common.TotalEnvelopesRepPerSec),
			util.GetFormattedUnit(totalLogRateUsed/8),
			len(TotalApps),
			len(InstanceMetricMap),
			util.GetFormattedUnit(totalMemoryAllocated),
			util.GetFormattedUnit(totalMemoryUsed))
	}

	if mainView != nil {
		mainView.Clear()
		common.MapLock.Lock()
		lineCounter := 0
		if common.ActiveView == common.AppInstanceView {
			mainView.Title = "Application Instances"
			_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %12s %5s %9s %7s %9s %6s %6s %9s %7s %-14s %9s %9s %-25s %-35s%s\n", conf.ColorYellow, "APP/INDEX", "LASTSEEN", "AGE", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "IP", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
			util.WriteToFile(fmt.Sprintf("InstanceMetricMap size: %d", len(InstanceMetricMap)))
			for _, pairlist := range sortedBy(InstanceMetricMap, common.ActiveSortDirection, activeInstancesSortField) {
				if passFilter(pairlist) {
					_, _ = fmt.Fprintf(mainView, "%s%-50s%s %s%5s%s %s%12s%s %s%5s%s %s%9s%s %s%7s%s %s%9s%s %s%6s%s %s%6s%s %s%9s%s %s%7s%s %s%-14s%s %s%9s%s %s%9s%s %s%-25s%s %s%-35s%s\n",
						appNameColor, fmt.Sprintf("%s/%s(%d)", util.TruncateString(pairlist.Value.AppName, 45), pairlist.Value.AppIndex, AppInstanceCounters[pairlist.Value.AppGuid].Count), conf.ColorReset,
						common.LastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), conf.ColorReset,
						common.AgeColor, util.GetFormattedElapsedTime(pairlist.Value.Tags[metricAge]), conf.ColorReset,
						cpuPercColor, util.GetFormattedUnit(pairlist.Value.Tags[MetricCpu]), conf.ColorReset,
						cpuTotColor, util.GetFormattedUnit(pairlist.Value.CpuTot), conf.ColorReset,
						memoryColor, util.GetFormattedUnit(pairlist.Value.Tags[metricMemory]), conf.ColorReset,
						memoryLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[metricMemoryQuota]), conf.ColorReset,
						diskColor, util.GetFormattedUnit(pairlist.Value.Tags[metricDisk]), conf.ColorReset,
						logRateColor, util.GetFormattedUnit(pairlist.Value.Tags[metricLogRate]), conf.ColorReset,
						logRateLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[metricLogRateLimit]), conf.ColorReset,
						entColor, util.GetFormattedUnit(pairlist.Value.Tags[metricCpuEntitlement]), conf.ColorReset,
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

		if common.ActiveView == common.AppView {
			mainView.Title = "Applications"
			_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %3s %4s %7s %8s %9s %5s %5s %9s %8s %7s %8s %-25s %-35s%s\n", conf.ColorYellow, "APP", "LASTSEEN", "IX", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
			for _, pairlist := range sortedBy(AppMetricMap, common.ActiveSortDirection, activeAppsSortField) {
				if passFilter(pairlist) {
					_, _ = fmt.Fprintf(mainView, "%s%-50s%s %s%5s%s %s%3d%s %s%4s%s %s%7s%s %s%8s%s %s%9s%s %s%5s%s %s%5s%s %s%9s%s %s%8s%s %s%7s%s %s%8s%s %s%-25s%s %s%-35s%s\n",
						appNameColor, fmt.Sprintf("%s", util.TruncateString(pairlist.Value.AppName, 45)), conf.ColorReset,
						common.LastSeenColor, util.GetFormattedElapsedTime(float64(time.Since(pairlist.Value.LastSeen).Nanoseconds())), conf.ColorReset,
						ixColor, pairlist.Value.IxCount, conf.ColorReset,
						cpuPercColor, util.GetFormattedUnit(pairlist.Value.Tags[MetricCpu]), conf.ColorReset,
						cpuTotColor, util.GetFormattedUnit(pairlist.Value.CpuTot), conf.ColorReset,
						memoryColor, util.GetFormattedUnit(pairlist.Value.Tags[metricMemory]), conf.ColorReset,
						memoryLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[metricMemoryQuota]), conf.ColorReset,
						diskColor, util.GetFormattedUnit(pairlist.Value.Tags[metricDisk]), conf.ColorReset,
						logRateColor, util.GetFormattedUnit(pairlist.Value.Tags[metricLogRate]), conf.ColorReset,
						logRateLimitColor, util.GetFormattedUnit(pairlist.Value.Tags[metricLogRateLimit]), conf.ColorReset,
						entColor, util.GetFormattedUnit(pairlist.Value.Tags[metricCpuEntitlement]), conf.ColorReset,
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
	}
	common.MapLock.Unlock()
}

func mkEvtHandler(ch rune) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if activeInstancesSortField == sortByAppName || activeAppsSortField == sortByAppName {
			if ch == rune(gocui.KeyBackspace) {
				if len(common.FilterStrings[filterFieldAppName]) > 0 {
					common.FilterStrings[filterFieldAppName] = common.FilterStrings[filterFieldAppName][:len(common.FilterStrings[filterFieldAppName])-1]
					_ = v.SetCursor(len(common.FilterStrings[filterFieldAppName])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				common.FilterStrings[filterFieldAppName] = common.FilterStrings[filterFieldAppName] + string(ch)
			}
		}
		if activeInstancesSortField == sortBySpace || activeAppsSortField == sortBySpace {
			if ch == rune(gocui.KeyBackspace) {
				if len(common.FilterStrings[filterFieldSpace]) > 0 {
					common.FilterStrings[filterFieldSpace] = common.FilterStrings[filterFieldSpace][:len(common.FilterStrings[filterFieldSpace])-1]
					_ = v.SetCursor(len(common.FilterStrings[filterFieldSpace])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				common.FilterStrings[filterFieldSpace] = common.FilterStrings[filterFieldSpace] + string(ch)
			}
		}
		if activeInstancesSortField == sortByOrg || activeAppsSortField == sortByOrg {
			if ch == rune(gocui.KeyBackspace) {
				if len(common.FilterStrings[filterFieldOrg]) > 0 {
					common.FilterStrings[filterFieldOrg] = common.FilterStrings[filterFieldOrg][:len(common.FilterStrings[filterFieldOrg])-1]
					_ = v.SetCursor(len(common.FilterStrings[filterFieldOrg])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				common.FilterStrings[filterFieldOrg] = common.FilterStrings[filterFieldOrg] + string(ch)
			}
		}
		return nil
	}
}

// updateAppMetrics - Populate the AppMetricMap with the latest instance metrics. */
func updateAppMetrics(instanceMetric *AppOrInstanceMetric) {
	var appMetric AppOrInstanceMetric
	var found bool
	if appMetric, found = AppMetricMap[instanceMetric.AppGuid]; !found {
		appMetric = AppOrInstanceMetric{
			LastSeen:  instanceMetric.LastSeen,
			AppName:   instanceMetric.AppName,
			AppGuid:   instanceMetric.AppGuid,
			IxCount:   1,
			SpaceName: instanceMetric.SpaceName,
			OrgName:   instanceMetric.OrgName,
			CpuTot:    instanceMetric.CpuTot,
			LogRtr:    instanceMetric.LogRtr,
			LogRep:    instanceMetric.LogRep,
			Tags:      make(map[string]float64),
		}
		for _, metricName := range MetricNames {
			appMetric.Tags[metricName] = instanceMetric.Tags[metricName]
		}
	} else {
		appMetric.LastSeen = instanceMetric.LastSeen
		appMetric.IxCount++
		appMetric.CpuTot += instanceMetric.CpuTot
		appMetric.LogRtr += instanceMetric.LogRtr
		appMetric.LogRep += instanceMetric.LogRep
		for _, metricName := range MetricNames {
			appMetric.Tags[metricName] += instanceMetric.Tags[metricName]
		}
	}
	AppMetricMap[instanceMetric.AppGuid] = appMetric
}
