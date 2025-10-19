package errco

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
)

// DebugLvl specify the level of debugging
// (start with LVL_3 to log config load errors)
var DebugLvl LogLvl = LVL_3

type MshLog struct {
	Ori LogOri        // log origin function
	Typ LogTyp        // log type
	Lvl LogLvl        // log debug level
	Cod LogCod        // log code
	Mex string        // log string
	Arg []interface{} // log args
}

type LogOri string
type LogTyp string
type LogLvl int
type LogCod int

// COLOR_GRAY is "\033[1m\033[30m" instead of "\033[1;30m"
// because `log.SetOutput(l.Stdout())` makes the color disappear for the second case
// https://github.com/gekware/minecraft-server-hibernation/blob/92607c76d9c9f872153578a612e88a5147a663ee/lib/input/input.go#L44

const (
	COLOR_RESET  = "\033[0m"
	COLOR_GRAY   = "\033[1m\033[30m"
	COLOR_RED    = "\033[31m"
	COLOR_GREEN  = "\033[32m"
	COLOR_YELLOW = "\033[33m"
	COLOR_BLUE   = "\033[34m"
	COLOR_PURPLE = "\033[35m"
	COLOR_CYAN   = "\033[36m"
)

var (
	userLogPrefix string
	userLogQuiet  bool
)

func init() {
	// disable log.Ldate and log.Ltime flags: will be set manually
	log.SetFlags(0)
}

// ConfigureUserLogging sets the plain output prefix and quiet mode flag used by Log.
func ConfigureUserLogging(prefix string, quiet bool) {
	userLogPrefix = prefix
	userLogQuiet = quiet
}

// NewLog returns a new msh log object.
//
// When a function fails and returns msh log using NewLog, msh log type must be TYPE_ERR or TYPE_WAR.
// Find bad usage with reg exp: `return (.*)NewLog\((.*)TYPE_(?!ERR|WAR)`
//
// To create a msh log and print it immediately you must use NewLogln()
// If you really want to use NewLog(), use NewLog().Log(false)
// Find bad usage with reg exp: `NewLog\((.*)\).Log\(true`
func NewLog(t LogTyp, l LogLvl, c LogCod, m string, a ...interface{}) *MshLog {
	logMsh := &MshLog{Trace(2), t, l, c, m, a}
	return logMsh
}

// NewLogln prints to terminal msh log struct and returns a new msh log struct.
//
// When a function fails it should not return a msh log struct using NewLogln
// (there is the risk of printing 2 times the same msh log)
// the parent function should handle the logging of msh log struct
// Find bad usage with reg exp: `return (.*)NewLogln\(`
func NewLogln(t LogTyp, l LogLvl, c LogCod, m string, a ...interface{}) *MshLog {
	logMsh := &MshLog{Trace(2), t, l, c, m, a}
	// trace was just set, no need to set it again
	// it would also be wrong:
	// 1) example()               -> Log() -> trace(2) : example
	// 2) example() -> NewLogln() -> trace(2)          : example
	//                            \> Log() -> trace(2) : NewLogln (!)
	logMsh.Log(false)
	return logMsh
}

// Log prints to terminal msh log struct.
//
// if tracing is set to true, Log() will add the caller function to the msh log trace
//
// returns the original log for convenience.
// returns nil if msh log struct is nil.
func (logMsh *MshLog) Log(tracing bool) *MshLog {
	// return immediately if original log is nil
	if logMsh == nil {
		return nil
	}

	// ------- operations on original log -------

	// add trace if requested
	if tracing {
		logMsh.Ori = Trace(2) + LogOri(" -> ") + logMsh.Ori
	}

	// return original log if log level is not high enough
	if logMsh.Lvl > DebugLvl {
		return logMsh
	}

	// make a copy of original log
	logMod := *logMsh

	if printPlainLog(&logMod) {
		return logMsh
	}

	// -------- operations on copied log --------

	var (
		typ string // log line: type    of log
		ori string // log line: origin  of log
		mex string // log line: message of log
		cod string // log line: code    of log
	)

	// set mex string depending on logMod level
	switch logMod.Lvl {
	case LVL_0: // make important logs more visible
		logMod.Mex = COLOR_CYAN + logMod.Mex + COLOR_RESET
	}

	// set typ, ori, mex, cod strings depending on logMod type
	switch logMod.Typ {
	case TYPE_INF:
		typ = fmt.Sprintf("%s%-6s%s", COLOR_BLUE, string(logMod.Typ), COLOR_RESET)
		ori = "\x00"
		mex = fmt.Sprintf(logMod.Mex, logMod.Arg...)
		cod = "\x00"
	case TYPE_SER:
		typ = fmt.Sprintf("%s%-6s%s", COLOR_GRAY, string(logMod.Typ), COLOR_RESET)
		ori = "\x00"
		mex = StringGraphic(fmt.Sprintf(logMod.Mex, logMod.Arg...)) // preserve original ANSI colors from minecraft server
		cod = "\x00"
	case TYPE_BYT:
		typ = fmt.Sprintf("%s%-6s%s", COLOR_PURPLE, string(logMod.Typ), COLOR_RESET)
		ori = "\x00"
		mex = fmt.Sprintf(logMod.Mex, logMod.Arg...)
		cod = "\x00"
	case TYPE_WAR:
		typ = fmt.Sprintf("%s%-6s%s", COLOR_YELLOW, string(logMod.Typ), COLOR_RESET)
		ori = fmt.Sprintf("%s%s:%s ", COLOR_YELLOW, logMod.Ori, COLOR_RESET)
		mex = fmt.Sprintf(logMod.Mex, logMod.Arg...)
		cod = fmt.Sprintf(" [%06x]", logMod.Cod)
	case TYPE_ERR:
		typ = fmt.Sprintf("%s%-6s%s", COLOR_RED, string(logMod.Typ), COLOR_RESET)
		ori = fmt.Sprintf("%s%s:%s ", COLOR_YELLOW, logMod.Ori, COLOR_RESET)
		mex = fmt.Sprintf(logMod.Mex, logMod.Arg...)
		cod = fmt.Sprintf(" [%06x]", logMod.Cod)
	}

	log.Printf("%s [%s%-4s] %s%s%s\n",
		time.Now().Format("2006/01/02 15:04:05.000"),
		typ,
		strings.Repeat("≡", 4-int(logMod.Lvl)),
		ori,
		mex,
		cod)

	// return original log
	return logMsh
}

func printPlainLog(logMod *MshLog) bool {
	if userLogPrefix == "" {
		return false
	}

	message := fmt.Sprintf(logMod.Mex, logMod.Arg...)

	switch logMod.Typ {
	case TYPE_INF:
		if userLogQuiet && logMod.Lvl > LVL_1 {
			return true
		}
		printPrefixed(message)
	case TYPE_BYT:
		if userLogQuiet {
			return true
		}
		printPrefixed(message)
	case TYPE_SER:
		printPrefixed(StringGraphic(message))
	case TYPE_WAR:
		printPrefixed(fmt.Sprintf("[WARN] %s", message))
	case TYPE_ERR:
		printPrefixed(fmt.Sprintf("[ERROR] %s", message))
	default:
		printPrefixed(message)
	}

	return true
}

func printPrefixed(message string) {
	if strings.HasSuffix(message, "\n") {
		fmt.Printf("[%s] %s", userLogPrefix, message)
	} else {
		fmt.Printf("[%s] %s\n", userLogPrefix, message)
	}
}

// AddTrace adds the caller function to the msh log trace
func (log *MshLog) AddTrace() *MshLog {
	// return original log if it's nil
	if log == nil {
		return log
	}

	log.Ori = Trace(2) + LogOri(" -> ") + log.Ori

	return log
}

// Trace returns the parent^(skip) function name
//
// skip == 2: example() -> NewLog() -> trace(): example
func Trace(skip int) LogOri {
	var o string = "?"

	if pc, _, _, ok := runtime.Caller(skip); !ok {
	} else if f := runtime.FuncForPC(pc); f == nil {
	} else {
		fn := f.Name()
		o = fn[strings.LastIndex(fn, ".")+1:]
	}

	return LogOri(o)
}
