package common

import (
	"errors"
	"fmt"
	"github.com/awesome-gocui/gocui"
	"github.com/metskem/rommel/MiniTopPlugin/util"
	"sort"
	"sync"
	"time"
)

const (
	FilterFieldIP int = iota
	FilterFieldJob
	FilterFieldAppName
	FilterFieldOrg
	FilterFieldSpace
)
const (
	AppInstanceView int = iota
	AppView
	VMView
	colorReset   = "\u001B[0m"
	colorReverse = "\u001B[34;7m"
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
	ShowToggleView          = false
	StartTime               = time.Now()
	FilterStrings           = make(map[int]string)
	ActiveSortDirection     = true
	ActiveView              = VMView
	ViewToggled             bool
	currentTogglePosition   int
	lines                   = make(map[int][]string)
)

func SetKeyBindings(gui *gocui.Gui) {
	_ = gui.SetKeybinding("", 'h', gocui.ModNone, help)
	_ = gui.SetKeybinding("", '?', gocui.ModNone, help)
	_ = gui.SetKeybinding("", 'q', gocui.ModNone, quit)
	_ = gui.SetKeybinding("", 't', gocui.ModNone, SetShowToggleView)
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

func help(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	ShowHelp = true
	return nil
}

func SetShowToggleView(g *gocui.Gui, v *gocui.View) error {
	_ = g // get rid of compiler warning
	_ = v // get rid of compiler warning
	ShowToggleView = true
	return nil
}

func ShowToggleViewLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if _, err := g.SetView("ToggleView", maxX/2-5, maxY/2-2, maxX/2+15, maxY/2+2, byte(0)); err != nil &&
		!errors.Is(err, gocui.ErrUnknownView) {
		return err
	} else {
		_ = g.SetKeybinding("ToggleView", gocui.KeyArrowDown, gocui.ModNone, arrowDown)
		_ = g.SetKeybinding("ToggleView", gocui.KeyArrowUp, gocui.ModNone, arrowUp)
		_ = g.SetKeybinding("ToggleView", gocui.KeyEnter, gocui.ModNone, enterToggle)
		if toggleView, err := g.SetCurrentView("ToggleView"); err != nil {
			util.WriteToFile(fmt.Sprintf("Error setting current view: %v", err))
		} else {
			lines[0] = []string{"", "VM View", ""}
			lines[1] = []string{"", "Application View", ""}
			lines[2] = []string{"", "App Instance View", ""}

			for i := 0; i < len(lines); i++ {
				if i == currentTogglePosition {
					lines[i] = []string{colorReverse, lines[i][1], colorReset}
				}
			}

			toggleView.Clear()
			toggleView.Title = "ToggleView"
			keys := make([]int, 0, len(lines))
			for k := range lines {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			for _, k := range keys {
				line := lines[len(keys)-k-1]
				_, _ = fmt.Fprintln(toggleView, fmt.Sprintf("%s%s%s", line[0], line[1], line[2]))
			}
		}
	}
	return nil
}

func arrowDown(g *gocui.Gui, v *gocui.View) error {
	if currentTogglePosition > 0 {
		currentTogglePosition -= 1
	}
	util.WriteToFile(fmt.Sprintf("Toggle arrowDown, currentTogglePostion=%d", currentTogglePosition))
	return nil
}

func arrowUp(g *gocui.Gui, v *gocui.View) error {
	if currentTogglePosition < 2 {
		currentTogglePosition += 1
	}
	util.WriteToFile(fmt.Sprintf("Toggle arrowUp, currentTogglePostion=%d", currentTogglePosition))
	return nil
}

func enterToggle(g *gocui.Gui, v *gocui.View) error {
	util.WriteToFile(fmt.Sprintf("Enter key pressed, currentSelection: %d", currentTogglePosition))
	_ = g.DeleteView("ToggleView")
	switch currentTogglePosition {
	case 0:
		_, _ = g.SetCurrentView("VMView")
		ActiveView = VMView
		ViewToggled = true
	case 1:
		_, _ = g.SetCurrentView("AppView")
		ActiveView = AppView
		ViewToggled = true
	case 2:
		_, _ = g.SetCurrentView("AppInstanceView")
		ActiveView = AppInstanceView
		ViewToggled = true
	}

	ShowToggleView = false
	return nil
}

func handleEnter(g *gocui.Gui, v *gocui.View) error {
	_ = v // get rid of compiler warning
	ShowFilter = false
	ShowHelp = false
	_ = g.DeleteView("FilterView")
	_ = g.DeleteView("HelpView")
	ViewToggled = true
	return nil
}
