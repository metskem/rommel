package util

import (
	"fmt"
	"os"
	"time"
)

// GetFormattedUnit - Transform the input (integer) to a string formatted in K, M or G */
func GetFormattedUnit(unitValue float64) string {
	if unitValue == 0 {
		return "-"
	}
	unitValueInt := int(unitValue)
	if unitValueInt >= 10*1024*1024*1024 {
		return fmt.Sprintf("%dG", unitValueInt/1024/1024/1024)
	} else if unitValueInt >= 10*1024*1024 {
		return fmt.Sprintf("%dM", unitValueInt/1024/1024)
	} else if unitValueInt >= 10*1024 {
		return fmt.Sprintf("%dK", unitValueInt/1024)
	} else {
		return fmt.Sprintf("%d", unitValueInt)
	}
}

// GetFormattedElapsedTime - Transform the input (time in nanoseconds) to a string with number of days, hours, mins and secs, like "1d01h54m10s" */
func GetFormattedElapsedTime(timeInNanoSecs float64) string {
	if timeInNanoSecs == 0 {
		return "-"
	}
	timeInSecs := int64(timeInNanoSecs / 1e9)
	days := timeInSecs / 86400
	secsLeft := timeInSecs % 86400
	hours := secsLeft / 3600
	secsLeft = secsLeft % 3600
	mins := secsLeft / 60
	secs := secsLeft % 60
	if days > 0 {
		return fmt.Sprintf("%dd%02dh%02dm%02ds", days, hours, mins, secs)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", hours, mins, secs)
	} else if mins > 0 {
		return fmt.Sprintf("%dm%02ds", mins, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}

func WriteToFile(text string) {
	f, _ := os.OpenFile("/tmp/gocui.out", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(time.Now().Format(time.RFC3339) + " " + text + "\n")
}
