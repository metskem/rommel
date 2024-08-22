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
	g            *gocui.Gui
	appNameColor = conf.ColorWhite
	ageColor     = conf.ColorWhite
	cpuColor     = conf.ColorWhite
	memoryColor  = conf.ColorWhite
	diskColor    = conf.ColorWhite
	logRateColor = conf.ColorWhite
	entColor     = conf.ColorWhite
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

	_ = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	_ = g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, arrowRight)
	_ = g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, arrowLeft)
	_ = g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, arrowDownOrUp)
	_ = g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, arrowDownOrUp)
	_ = g.SetKeybinding("", 'n', gocui.ModNone, sortByAppName)
	_ = g.SetKeybinding("", 'a', gocui.ModNone, sortByAge)
	_ = g.SetKeybinding("", 'c', gocui.ModNone, sortByCpu)
	_ = g.SetKeybinding("", 'o', gocui.ModNone, sortByMemory)
	_ = g.SetKeybinding("", 'm', gocui.ModNone, sortByLogRate)
	_ = g.SetKeybinding("", 'd', gocui.ModNone, sortByDisk)
	_ = g.SetKeybinding("", 'e', gocui.ModNone, sortByEntitlement)

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
	if mainView, err = g.SetView("ApplicationView", 0, 0, maxX*9/10, maxY*9/10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		mainView.Title = "Application View"
		//mainView.Autoscroll = true
		//mainView.Editable = true
	}
	return nil
}

func refreshViewContent() {
	mainView.Clear()
	maxX, maxY := g.Size()
	if err := mainView.SetCursor(maxX/2-maxX/4, maxY/2-maxY/4); err != nil {
		util.WriteToFile("Error setting cursor: " + err.Error())
	}
	//yellow := color.New(color.Bold, color.FgYellow).SprintFunc()
	_, _ = fmt.Fprint(mainView, fmt.Sprintf("%s%-62s %15s %10s %6s %8s %6s %9s %-20s %-35s%s\n", conf.ColorYellow, "APP/INDEX", "AGE", "CPU", "MEMORY", "DISK", "LOGRATE", "CPU_ENT", "ORG", "SPACE", conf.ColorReset))
	conf.MapLock.Lock()

	for _, pairlist := range util.SortedBy(conf.MetricMap, util.ActiveSortDirection, util.ActiveSortField) {
		_, _ = fmt.Fprintf(mainView, "%s%-65s%s %s%12s%s %s%10.f%s %s%6s%s %s%8s%s %s%7s%s %s%9.f%s %s%-20s%s %s%-35s%s\n",
			appNameColor, pairlist.Value.AppName+"/"+pairlist.Value.AppIndex, conf.ColorReset,
			ageColor, util.GetFormattedElapsedTime(pairlist.Value.Values["container_age"]), conf.ColorReset,
			cpuColor, pairlist.Value.Values["cpu"], conf.ColorReset,
			memoryColor, util.GetFormattedUnit(pairlist.Value.Values["memory"]), conf.ColorReset,
			diskColor, util.GetFormattedUnit(pairlist.Value.Values["disk"]), conf.ColorReset,
			logRateColor, util.GetFormattedUnit(pairlist.Value.Values["log_rate"]), conf.ColorReset,
			entColor, pairlist.Value.Values["cpu_entitlement"], conf.ColorReset,
			orgColor, pairlist.Value.OrgName, conf.ColorReset,
			spaceColor, pairlist.Value.SpaceName, conf.ColorReset)
	}

	conf.MapLock.Unlock()
}

func quit(g *gocui.Gui, v *gocui.View) error {
	os.Exit(0)
	return gocui.ErrQuit
}
func sortByAge(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField == util.SortByAge {
		flipSortOrder()
	}
	util.ActiveSortField = util.SortByAge
	return nil
}
func sortByAppName(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField == util.SortByAppName {
		flipSortOrder()
	}
	util.ActiveSortField = util.SortByAge
	//appNameColor = color.New(color.FgBlue).SprintFunc()
	return nil
}
func sortByCpu(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField == util.SortByCpu {
		flipSortOrder()
	}
	util.ActiveSortField = util.SortByCpu
	cpuColor = conf.ColorBlue
	util.WriteToFile("Sorting by CPU")
	return nil
}
func sortByMemory(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField == util.SortByMemory {
		flipSortOrder()
	}
	util.ActiveSortField = util.SortByMemory
	return nil
}
func sortByDisk(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField == util.SortByDisk {
		flipSortOrder()
	}
	util.ActiveSortField = util.SortByDisk
	return nil
}
func sortByLogRate(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField == util.SortByLogRate {
		flipSortOrder()
	}
	util.ActiveSortField = util.SortByLogRate
	return nil
}
func sortByEntitlement(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField == util.SortByEntitlement {
		flipSortOrder()
	}
	util.ActiveSortField = util.SortByEntitlement
	return nil
}
func arrowRight(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField != util.SortBySpace {
		util.ActiveSortField++
	}
	colorSortedColumn()
	return nil
}
func arrowLeft(g *gocui.Gui, v *gocui.View) error {
	if util.ActiveSortField != util.SortByAppName {
		util.ActiveSortField--
	}
	colorSortedColumn()
	return nil
}
func arrowDownOrUp(g *gocui.Gui, v *gocui.View) error {
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
	cpuColor = conf.ColorWhite
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
	case util.SortByCpu:
		cpuColor = conf.ColorBlue
	case util.SortByMemory:
		memoryColor = conf.ColorBlue
	case util.SortByDisk:
		diskColor = conf.ColorBlue
	case util.SortByLogRate:
		logRateColor = conf.ColorBlue
	case util.SortByEntitlement:
		entColor = conf.ColorBlue
	case util.SortByOrg:
		orgColor = conf.ColorBlue
	case util.SortBySpace:
		spaceColor = conf.ColorBlue
	}
}
