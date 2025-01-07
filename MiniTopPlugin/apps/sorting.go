package apps

import (
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/common"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"github.com/metskem/rommel/MiniTopPlugin/util"
)

var (
	appNameColor      = conf.ColorWhite
	cpuPercColor      = conf.ColorWhite
	ixColor           = conf.ColorWhite
	cpuTotColor       = conf.ColorWhite
	memoryColor       = conf.ColorWhite
	memoryLimitColor  = conf.ColorWhite
	diskColor         = conf.ColorWhite
	logRateColor      = conf.ColorWhite
	logRateLimitColor = conf.ColorWhite
	entColor          = conf.ColorWhite
	logRepColor       = conf.ColorWhite
	logRtrColor       = conf.ColorWhite
	orgColor          = conf.ColorWhite
	spaceColor        = conf.ColorWhite
)

func spacePressed(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	common.FlipSortOrder()
	return nil
}

func arrowRight(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	if util.ActiveInstancesSortField != util.SortBySpace {
		util.ActiveInstancesSortField++
	}
	if util.ActiveAppsSortField != util.SortBySpace {
		util.ActiveAppsSortField++
	}
	// when in instance view mode, there is no Ix column, so skip it
	if conf.ActiveView == conf.AppInstanceView {
		if util.ActiveInstancesSortField == util.SortByIx {
			util.ActiveInstancesSortField++
		}
	}
	// when in app view mode, the Age and IP columns are not there, so skip them
	if conf.ActiveView == conf.AppView {
		if util.ActiveAppsSortField == util.SortByAge || util.ActiveAppsSortField == util.SortByIP {
			util.ActiveAppsSortField++
		}
	}
	colorSortedColumn()
	return nil
}

func arrowLeft(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	if util.ActiveInstancesSortField != util.SortByAppName {
		util.ActiveInstancesSortField--
	}
	if util.ActiveAppsSortField != util.SortByAppName {
		util.ActiveAppsSortField--
	}
	// when in instance view mode, there is no Ix column, so skip it
	if conf.ActiveView == conf.AppInstanceView {
		if util.ActiveInstancesSortField == util.SortByIx {
			util.ActiveInstancesSortField--
		}
	}
	// when in app view mode, the Age and IP columns are not there, so skip them
	if conf.ActiveView == conf.AppView {
		if util.ActiveAppsSortField == util.SortByAge || util.ActiveAppsSortField == util.SortByIP {
			util.ActiveAppsSortField--
		}
	}
	colorSortedColumn()
	return nil
}

func colorSortedColumn() {
	appNameColor = conf.ColorWhite
	common.LastSeenColor = conf.ColorWhite
	common.AgeColor = conf.ColorWhite
	cpuPercColor = conf.ColorWhite
	ixColor = conf.ColorWhite
	cpuTotColor = conf.ColorWhite
	memoryColor = conf.ColorWhite
	memoryLimitColor = conf.ColorWhite
	diskColor = conf.ColorWhite
	logRateColor = conf.ColorWhite
	logRateLimitColor = conf.ColorWhite
	entColor = conf.ColorWhite
	common.IPColor = conf.ColorWhite
	logRepColor = conf.ColorWhite
	logRtrColor = conf.ColorWhite
	orgColor = conf.ColorWhite
	spaceColor = conf.ColorWhite
	if conf.ActiveView == conf.AppInstanceView {
		switch util.ActiveInstancesSortField {
		case util.SortByAppName:
			appNameColor = conf.ColorBlue
		case util.SortByLastSeen:
			common.LastSeenColor = conf.ColorBlue
		case util.SortByAge:
			common.AgeColor = conf.ColorBlue
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
			common.IPColor = conf.ColorBlue
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
	if conf.ActiveView == conf.AppView {
		switch util.ActiveAppsSortField {
		case util.SortByAppName:
			appNameColor = conf.ColorBlue
		case util.SortByLastSeen:
			common.LastSeenColor = conf.ColorBlue
		case util.SortByAge:
			common.AgeColor = conf.ColorBlue
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
			common.IPColor = conf.ColorBlue
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
}
