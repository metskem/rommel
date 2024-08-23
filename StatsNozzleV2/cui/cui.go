package cui

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/metskem/rommel/StatsNozzleV2/conf"
	"github.com/metskem/rommel/StatsNozzleV2/util"
	"log"
	"os"
	"time"
)

var (
	mainView     *gocui.View
	summaryView  *gocui.View
	g            *gocui.Gui
	appNameColor = conf.ColorWhite
	ageColor     = conf.ColorWhite
	cpuPercColor = conf.ColorWhite
	cpuTotColor  = conf.ColorWhite
	memoryColor  = conf.ColorWhite
	diskColor    = conf.ColorWhite
	logRateColor = conf.ColorWhite
	entColor     = conf.ColorWhite
	logRepColor  = conf.ColorWhite
	logRtrColor  = conf.ColorWhite
	orgColor     = conf.ColorWhite
	spaceColor   = conf.ColorWhite
)

func Start() {
	var err error
	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.InputEsc = true

	//_ = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	_ = g.SetKeybinding("", 'q', gocui.ModNone, quit)
	_ = g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, arrowRight)
	_ = g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, arrowLeft)
	_ = g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, arrowDownOrUp)
	_ = g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, arrowDownOrUp)

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
		summaryView.Title = "Summary"
	}
	if mainView, err = g.SetView("ApplicationView", 0, 5, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		mainView.Title = "Application Instances"
	}
	return nil
}

func refreshViewContent() {
	//maxX, maxY := g.Size()

	summaryView.Clear()
	_, _ = fmt.Fprintf(summaryView, "Target: %s\nTotal events: %d\nTotal App Instances: %d", conf.ApiAddr, conf.TotalEnvelopes, len(conf.MetricMap))

	mainView.Clear()
	//if err := mainView.SetCursor(maxX/2-maxX/4, maxY/2-maxY/4); err != nil {
	//	util.WriteToFile("Error setting cursor: " + err.Error())
	//}
	_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-62s %15s %10s %12s %6s %8s %6s %9s %9s %9s %-25s %-35s%s\n", conf.ColorYellow, "APP/INDEX", "AGE", "CPU%", "CPUTOT", "MEMORY", "DISK", "LOGRATE", "CPU_ENT", "LOG_REP", "LOG_RTR", "ORG", "SPACE", conf.ColorReset))
	conf.MapLock.Lock()

	for _, pairlist := range util.SortedBy(conf.MetricMap, util.ActiveSortDirection, util.ActiveSortField) {
		_, _ = fmt.Fprintf(mainView, "%s%-65s%s %s%12s%s %s%10s%s %s%12s%s %s%6s%s %s%8s%s %s%7s%s %s%9s%s %s%9s%s %s%9s%s %s%-25s%s %s%-35s%s\n",
			appNameColor, pairlist.Value.AppName+"/"+pairlist.Value.AppIndex, conf.ColorReset,
			ageColor, util.GetFormattedElapsedTime(pairlist.Value.Values["container_age"]), conf.ColorReset,
			cpuPercColor, util.GetFormattedUnit(pairlist.Value.Values["cpu"]), conf.ColorReset,
			cpuTotColor, util.GetFormattedUnit(pairlist.Value.CpuTot), conf.ColorReset,
			memoryColor, util.GetFormattedUnit(pairlist.Value.Values["memory"]), conf.ColorReset,
			diskColor, util.GetFormattedUnit(pairlist.Value.Values["disk"]), conf.ColorReset,
			logRateColor, util.GetFormattedUnit(pairlist.Value.Values["log_rate"]), conf.ColorReset,
			entColor, util.GetFormattedUnit(pairlist.Value.Values["cpu_entitlement"]), conf.ColorReset,
			logRepColor, util.GetFormattedUnit(pairlist.Value.LogRep), conf.ColorReset,
			logRtrColor, util.GetFormattedUnit(pairlist.Value.LogRtr), conf.ColorReset,
			orgColor, pairlist.Value.OrgName, conf.ColorReset,
			spaceColor, pairlist.Value.SpaceName, conf.ColorReset)
	}

	conf.MapLock.Unlock()
}

func quit(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	os.Exit(0)
	return gocui.ErrQuit
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
func arrowDownOrUp(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	flipSortOrder()
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
	ageColor = conf.ColorWhite
	cpuPercColor = conf.ColorWhite
	cpuTotColor = conf.ColorWhite
	memoryColor = conf.ColorWhite
	diskColor = conf.ColorWhite
	logRateColor = conf.ColorWhite
	entColor = conf.ColorWhite
	orgColor = conf.ColorWhite
	spaceColor = conf.ColorWhite
	switch util.ActiveSortField {
	case util.SortByAppName:
		appNameColor = conf.ColorBlue
	case util.SortByAge:
		ageColor = conf.ColorBlue
	case util.SortByCpuPerc:
		cpuPercColor = conf.ColorBlue
	case util.SortByCpuTot:
		cpuTotColor = conf.ColorBlue
	case util.SortByMemory:
		memoryColor = conf.ColorBlue
	case util.SortByDisk:
		diskColor = conf.ColorBlue
	case util.SortByLogRate:
		logRateColor = conf.ColorBlue
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
