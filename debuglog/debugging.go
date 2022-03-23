package debuglog

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func findFunc(f *runtime.Frame) (string, string) {
	s := strings.Split(f.Function, ".")
	funcname := s[len(s)-1]

	return funcname, ""
}

var fieldSeq = map[string]int{
	"time":  0,
	"level": 1,
	"func":  2,
}

func sortCustom(fields []string) {
	sort.Slice(fields, func(i, j int) bool {
		if fields[i] == "msg" {
			return false
		}
		if iIdx, oki := fieldSeq[fields[i]]; oki {
			if jIdx, okj := fieldSeq[fields[j]]; okj {
				return iIdx < jIdx
			}
			return true
		}
		return false
	})
}

func DebugLogInit(logname string) *logrus.Logger {

	cwflogrus := logrus.New()
	wafFormatter := &logrus.TextFormatter{
		DisableColors:    true,
		CallerPrettyfier: findFunc,
		ForceQuote:       true,
		DisableSorting:   false,
		SortingFunc:      sortCustom,
	}

	cwflogrus.SetReportCaller(true)
	cwflogrus.SetFormatter(wafFormatter)

	// path
	// baselogpath := os.Getenv("DEBUG_DIR")

	baselogpath, err := os.Getwd()
	if err != nil {
		fmt.Println("Cannot get CurrentDirectory")
	}
	if len(baselogpath) == 0 {
		// baselogpath = "/monitorapp/debug"
		baselogpath = "/Users/byeoungwoolee/go/src/apilimit"
	}

	// WAF LOG OUTPUT SETTING
	cwflogpath := fmt.Sprintf("%s/%s.%d.log", baselogpath, logname, os.Getpid())
	cwfLogOutput := lumberjack.Logger{
		Filename:   cwflogpath,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     3,
		Compress:   false,
	}

	cwflogrus.SetOutput(&cwfLogOutput)
	return cwflogrus
}
