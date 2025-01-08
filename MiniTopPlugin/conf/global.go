package conf

import (
	"sync"
	"time"
)

const (
	TagIp       = "ip"
	ColorReset  = "\033[0m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[36m"
	ColorWhite  = "\033[97m"
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
	IntervalSecs            = 1
)
