package common

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"github.com/metskem/rommel/MiniTopPlugin/util"
)

var (
	ActiveSortDirection bool
)

func SetKeyBindings(gui *gocui.Gui) {
	_ = gui.SetKeybinding("", 'h', gocui.ModNone, help)
	_ = gui.SetKeybinding("", '?', gocui.ModNone, help)
	_ = gui.SetKeybinding("", 'q', gocui.ModNone, quit)
	_ = gui.SetKeybinding("HelpView", gocui.KeyEnter, gocui.ModNone, handleEnter)
	_ = gui.SetKeybinding("FilterView", gocui.KeyEnter, gocui.ModNone, handleEnter)
}

func quit(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	//os.Exit(0)
	return gocui.ErrQuit
}

func SpacePressed(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	FlipSortOrder()
	return nil
}

func FlipSortOrder() {
	util.WriteToFile("apps.FlipSortOrder")
	if conf.ActiveView == conf.AppView || conf.ActiveView == conf.AppInstanceView {
		if ActiveSortDirection == true {
			ActiveSortDirection = false
		} else {
			ActiveSortDirection = true
		}
		util.WriteToFile(fmt.Sprintf("Apps ActiveSortDirection: %t", ActiveSortDirection))
	}
	if conf.ActiveView == conf.VMView {
		if ActiveSortDirection == true {
			ActiveSortDirection = false
		} else {
			ActiveSortDirection = true
		}
		util.WriteToFile(fmt.Sprintf("VMs ActiveSortDirection: %t", ActiveSortDirection))
	}
}

func ShowFilterView(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	conf.ShowFilter = true
	return nil
}

func help(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	conf.ShowHelp = true
	return nil
}

func ToggleView(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	if conf.ActiveView == conf.AppInstanceView {
		conf.ActiveView = conf.AppView
	} else {
		if conf.ActiveView == conf.AppView {
			conf.ActiveView = conf.VMView
		} else {
			if conf.ActiveView == conf.VMView {
				conf.ActiveView = conf.AppInstanceView
			}
		}
	}
	util.WriteToFile(fmt.Sprintf("ActiveView: %d", conf.ActiveView))
	return nil
}

func handleEnter(g *gocui.Gui, v *gocui.View) error {
	_ = v // get rid of compiler warning
	conf.ShowFilter = false
	conf.ShowHelp = false
	_ = g.DeleteView("FilterView")
	_ = g.DeleteView("HelpView")
	_, _ = g.SetCurrentView("ApplicationView")
	return nil
}
