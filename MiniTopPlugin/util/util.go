package util

import (
	"fmt"
	"github.com/metskem/rommel/MiniTopPlugin/conf"
	"os"
	"regexp"
	"time"
)

var logFile *os.File

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
	var err error
	if logFile == nil {
		if logFile, err = os.OpenFile("/tmp/gocui.out", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644); err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			os.Exit(1)
		}
	}
	//defer func() { _ = logFile.Close() }()
	_, _ = logFile.WriteString(time.Now().Format(time.RFC3339) + " " + text + "\n")
}

// UpdateAppMetrics - Populate the AppMetricMap with the latest instance metrics. */
func UpdateAppMetrics(instanceMetric *conf.AppOrInstanceMetric) {
	var appMetric conf.AppOrInstanceMetric
	var found bool
	if appMetric, found = conf.AppMetricMap[instanceMetric.AppGuid]; !found {
		appMetric = conf.AppOrInstanceMetric{
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
		for _, metricName := range conf.MetricNames {
			appMetric.Tags[metricName] = instanceMetric.Tags[metricName]
		}
	} else {
		appMetric.LastSeen = instanceMetric.LastSeen
		appMetric.IxCount++
		appMetric.CpuTot += instanceMetric.CpuTot
		appMetric.LogRtr += instanceMetric.LogRtr
		appMetric.LogRep += instanceMetric.LogRep
		for _, metricName := range conf.MetricNames {
			appMetric.Tags[metricName] += instanceMetric.Tags[metricName]
		}
	}
	conf.AppMetricMap[instanceMetric.AppGuid] = appMetric
}

func TruncateString(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}

func PassFilter(pairList Pair) bool {
	passFilter := true
	filterRegex := regexp.MustCompile(conf.FilterStrings[conf.FilterFieldAppName])
	if !(conf.FilterStrings[conf.FilterFieldAppName] == "") && !filterRegex.MatchString(pairList.Value.AppName) {
		passFilter = false
	}
	filterRegex = regexp.MustCompile(conf.FilterStrings[conf.FilterFieldSpace])
	if !(conf.FilterStrings[conf.FilterFieldSpace] == "") && !filterRegex.MatchString(pairList.Value.SpaceName) {
		passFilter = false
	}
	filterRegex = regexp.MustCompile(conf.FilterStrings[conf.FilterFieldOrg])
	if !(conf.FilterStrings[conf.FilterFieldOrg] == "") && !filterRegex.MatchString(pairList.Value.OrgName) {
		passFilter = false
	}
	return passFilter
}
