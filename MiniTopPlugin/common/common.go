package common

import (
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"sync"
	"time"
)

const (
	AppInstanceView int = iota
	AppView
	VMView
)

var (
	MapLock                 sync.Mutex
	TotalEnvelopes          float64
	TotalEnvelopesPerSec    float64
	TotalEnvelopesRep       float64
	TotalEnvelopesRepPerSec float64
	TotalEnvelopesRtr       float64
	TotalEnvelopesRtrPerSec float64
	ShowFilter              = false
	ShowHelp                = false
	StartTime               = time.Now()
	FilterStrings           = make(map[int]string)
	ActiveSortDirection     = true
	ActiveView              = VMView
	ViewToggled             bool
)

func SetKeyBindings(gui *gocui.Gui) {
	_ = gui.SetKeybinding("", 'h', gocui.ModNone, help)
	_ = gui.SetKeybinding("", '?', gocui.ModNone, help)
	_ = gui.SetKeybinding("", 'q', gocui.ModNone, quit)
	_ = gui.SetKeybinding("", 't', gocui.ModNone, toggleView)
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
	if ActiveView == AppView || ActiveView == AppInstanceView {
		if ActiveSortDirection == true {
			ActiveSortDirection = false
		} else {
			ActiveSortDirection = true
		}
	}
	if ActiveView == VMView {
		if ActiveSortDirection == true {
			ActiveSortDirection = false
		} else {
			ActiveSortDirection = true
		}
	}
}

func ShowFilterView(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	ShowFilter = true
	return nil
}

func help(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	ShowHelp = true
	return nil
}

func toggleView(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	ViewToggled = true
	if ActiveView == AppInstanceView {
		ActiveView = AppView
	} else {
		if ActiveView == AppView {
			ActiveView = VMView
		} else {
			if ActiveView == VMView {
				ActiveView = AppInstanceView
			}
		}
	}
	util.WriteToFile(fmt.Sprintf("ActiveView: %d", ActiveView))
	return nil
}

func handleEnter(g *gocui.Gui, v *gocui.View) error {
	_ = v // get rid of compiler warning
	ShowFilter = false
	ShowHelp = false
	_ = g.DeleteView("FilterView")
	_ = g.DeleteView("HelpView")
	_, _ = g.SetCurrentView("ApplicationView")
	return nil
}
