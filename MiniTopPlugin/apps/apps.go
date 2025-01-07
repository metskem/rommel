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
)

var (
	mainView             *gocui.View
	summaryView          *gocui.View
	AppMetricMap         map[string]AppOrInstanceMetric         // map key is app-guid
	InstanceMetricMap    = make(map[string]AppOrInstanceMetric) // map key is app-guid/index
	AppInstanceCounters  = make(map[string]AppInstanceCounter)  // here we keep the highest instance index for each app
	MetricCpu            = "cpu"
	metricAge            = "container_age"
	metricCpuEntitlement = "cpu_entitlement"
	metricDisk           = "disk"
	metricMemory         = "memory"
	metricMemoryQuota    = "memory_quota"
	metricLogRate        = "log_rate"
	metricLogRateLimit   = "log_rate_limit"
	TagOrgName           = "organization_name"
	TagSpaceName         = "space_name"
	TagAppName           = "app_name"
	TagAppId             = "app_id"
	TagAppInstanceId     = "instance_id" // use this for app index
	TagOrigin            = "origin"
	TagOriginValueRep    = "rep"
	TagOriginValueRtr    = "gorouter"

	MetricNames = []string{MetricCpu, metricAge, metricCpuEntitlement, metricDisk, metricMemory, metricMemoryQuota, metricLogRate, metricLogRateLimit}
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
	return Layout(g)
}

func ShowView(gui *gocui.Gui) {
	colorSortedColumn()
	totalEnvelopesPrev := conf.TotalEnvelopes
	totalEnvelopesRepPrev := conf.TotalEnvelopesRep
	totalEnvelopesRtrPrev := conf.TotalEnvelopesRtr

	// update memory summaries
	var totalMemUsed, totalMemAllocated, totalLogRateUsed float64
	conf.MapLock.Lock()
	AppMetricMap = make(map[string]AppOrInstanceMetric)
	for _, metric := range InstanceMetricMap {
		totalMemUsed += metric.Tags[metricMemory]
		totalMemAllocated += metric.Tags[metricMemoryQuota]
		totalLogRateUsed += metric.Tags[metricLogRate]
		updateAppMetrics(&metric)
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
	conf.FilterStrings[filterFieldAppName] = ""
	conf.FilterStrings[filterFieldOrg] = ""
	conf.FilterStrings[filterFieldSpace] = ""
	return nil
}

func Layout(g *gocui.Gui) (err error) {
	if conf.ActiveView != conf.AppView && conf.ActiveView != conf.AppInstanceView {
		return nil
	}
	util.WriteToFile("APP/Instances layout")
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
			if activeAppsSortField == sortByAppName || activeInstancesSortField == sortByAppName {
				_, _ = fmt.Fprintln(v, " AppName")
				_, _ = fmt.Fprintln(v, conf.FilterStrings[filterFieldAppName])
			}
			if activeAppsSortField == sortBySpace || activeInstancesSortField == sortBySpace {
				_, _ = fmt.Fprintln(v, " Space")
				_, _ = fmt.Fprintln(v, conf.FilterStrings[filterFieldSpace])
			}
			if activeAppsSortField == sortByOrg || activeInstancesSortField == sortByOrg {
				_, _ = fmt.Fprintln(v, " Org")
				_, _ = fmt.Fprintln(v, conf.FilterStrings[filterFieldOrg])
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
	util.WriteToFile("App/Instances refreshViewContent")
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
		len(InstanceMetricMap),
		util.GetFormattedUnit(conf.TotalMemoryAllocated),
		util.GetFormattedUnit(conf.TotalMemoryUsed))

	mainView.Clear()
	conf.MapLock.Lock()
	lineCounter := 0
	if conf.ActiveView == conf.AppInstanceView {
		mainView.Title = "Application Instances"
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %12s %5s %9s %7s %9s %6s %6s %9s %7s %-14s %9s %9s %-25s %-35s%s\n", conf.ColorYellow, "APP/INDEX", "LASTSEEN", "AGE", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "IP", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
		for _, pairlist := range sortedBy(InstanceMetricMap, ActiveSortDirection, activeInstancesSortField) {
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
	if conf.ActiveView == conf.AppView {
		mainView.Title = "Applications"
		_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-47s %8s %3s %4s %7s %8s %9s %5s %5s %9s %8s %7s %8s %-25s %-35s%s\n", conf.ColorYellow, "APP", "LASTSEEN", "IX", "CPU%", "CPUTOT", "MEMORY", "MEM_QUOTA", "DISK", "LOGRT", "LOGRT_LIM", "CPU_ENT", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
		for _, pairlist := range sortedBy(AppMetricMap, ActiveSortDirection, activeAppsSortField) {
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
	conf.MapLock.Unlock()
}

func mkEvtHandler(ch rune) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if activeInstancesSortField == sortByAppName || activeAppsSortField == sortByAppName {
			if ch == rune(gocui.KeyBackspace) {
				if len(conf.FilterStrings[filterFieldAppName]) > 0 {
					conf.FilterStrings[filterFieldAppName] = conf.FilterStrings[filterFieldAppName][:len(conf.FilterStrings[filterFieldAppName])-1]
					_ = v.SetCursor(len(conf.FilterStrings[filterFieldAppName])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				conf.FilterStrings[filterFieldAppName] = conf.FilterStrings[filterFieldAppName] + string(ch)
			}
		}
		if activeInstancesSortField == sortBySpace || activeAppsSortField == sortBySpace {
			if ch == rune(gocui.KeyBackspace) {
				if len(conf.FilterStrings[filterFieldSpace]) > 0 {
					conf.FilterStrings[filterFieldSpace] = conf.FilterStrings[filterFieldSpace][:len(conf.FilterStrings[filterFieldSpace])-1]
					_ = v.SetCursor(len(conf.FilterStrings[filterFieldSpace])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				conf.FilterStrings[filterFieldSpace] = conf.FilterStrings[filterFieldSpace] + string(ch)
			}
		}
		if activeInstancesSortField == sortByOrg || activeAppsSortField == sortByOrg {
			if ch == rune(gocui.KeyBackspace) {
				if len(conf.FilterStrings[filterFieldOrg]) > 0 {
					conf.FilterStrings[filterFieldOrg] = conf.FilterStrings[filterFieldOrg][:len(conf.FilterStrings[filterFieldOrg])-1]
					_ = v.SetCursor(len(conf.FilterStrings[filterFieldOrg])+1, 1)
					v.EditDelete(true)
				}
				return nil
			} else {
				_, _ = fmt.Fprint(v, string(ch))
				conf.FilterStrings[filterFieldOrg] = conf.FilterStrings[filterFieldOrg] + string(ch)
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
