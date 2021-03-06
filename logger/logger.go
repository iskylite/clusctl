package logger

import (
	"fmt"
	"io"
	"log"
	"myclush/utils"
	"os"
)

const (
	DEBUG = iota
	INFO
	ERROR
)

var Logger *logger

func init() {
	Logger = newlogger(INFO)
}

type logger struct {
	level  int
	silent bool
	color  bool
	log    *log.Logger
}

func SetLogLevel(debug bool) {
	if debug {
		SetLevel(DEBUG)
	} else {
		SetSilent()
	}
}

func newlogger(level int) *logger {
	defaultLogger := log.Default()
	defaultLogger.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmsgprefix)
	return &logger{
		level:  level,
		silent: false,
		color:  true,
		log:    defaultLogger,
	}
}

func SetLevel(level int) {
	Logger.level = level
}

func SetSilent() {
	Logger.silent = true
}

func DisableColor() {
	Logger.color = false
}

func SetOutput(output io.Writer) {
	Logger.log.SetOutput(output)
}

func SetOutputFile(logFile string) (*os.File, error) {
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return f, err
	}
	SetOutput(f)
	return f, nil
}

func Debug(a ...interface{}) {
	if Logger.level == DEBUG {
		if Logger.silent {
			if Logger.log.Writer() != nil {
				fmt.Fprint(Logger.log.Writer(), a...)
			} else {
				fmt.Println(a...)
			}
		} else {
			lp := "DEBG "
			if Logger.color {
				lp = "\x1b[32mDEBG \x1b[0m"
			}
			v := append([]interface{}{lp}, a...)
			Logger.log.Output(2, fmt.Sprintln(v...))
		}
	}
}

func Debugf(format string, v ...interface{}) {
	if Logger.level == DEBUG {
		if Logger.silent {
			if Logger.log.Writer() != nil {
				fmt.Fprintf(Logger.log.Writer(), format, v...)
			} else {
				fmt.Printf(format, v...)
			}
		} else {
			lp := "DEBG "
			if Logger.color {
				lp = "\x1b[32mDEBG \x1b[0m"
			}
			levelFormat := fmt.Sprintf("%s %s", lp, format)
			Logger.log.Output(2, fmt.Sprintf(levelFormat, v...))
		}
	}
}
func Info(a ...interface{}) {
	if Logger.level <= INFO {
		if Logger.silent {
			if Logger.log.Writer() != nil {
				fmt.Fprint(Logger.log.Writer(), a...)
			} else {
				fmt.Println(a...)
			}
		} else {
			lp := "INFO "
			if Logger.color {
				lp = "\x1b[36mINFO \x1b[0m"
			}
			v := append([]interface{}{lp}, a...)
			Logger.log.Output(2, fmt.Sprintln(v...))
		}
	}
}

func Infof(format string, v ...interface{}) {
	if Logger.level <= INFO {
		if Logger.silent {
			if Logger.log.Writer() != nil {
				fmt.Fprintf(Logger.log.Writer(), format, v...)
			} else {
				fmt.Printf(format, v...)
			}
		} else {
			lp := "INFO "
			if Logger.color {
				lp = "\x1b[36mINFO \x1b[0m"
			}
			levelFormat := fmt.Sprintf("%s %s", lp, format)
			Logger.log.Output(2, fmt.Sprintf(levelFormat, v...))
		}
	}
}

func Error(a ...interface{}) {
	if Logger.level <= ERROR {
		if Logger.silent {
			if Logger.log.Writer() != nil {
				fmt.Fprint(Logger.log.Writer(), a...)
			} else {
				fmt.Println(a...)
			}
		} else {
			lp := "ERRO "
			if Logger.color {
				lp = "\x1b[31mERRO \x1b[0m"
			}
			v := append([]interface{}{lp}, a...)
			Logger.log.Output(2, fmt.Sprintln(v...))
		}
	}
}

func Errorf(format string, v ...interface{}) {
	if Logger.level <= ERROR {
		if Logger.silent {
			if Logger.log.Writer() != nil {
				fmt.Fprintf(Logger.log.Writer(), format, v...)
			} else {
				fmt.Printf(format, v...)
			}
		} else {
			lp := "ERRO "
			if Logger.color {
				lp = "\x1b[31mERRO \x1b[0m"
			}
			levelFormat := fmt.Sprintf("%s %s", lp, format)
			Logger.log.Output(2, fmt.Sprintf(levelFormat, v...))
		}
	}
}

const (
	Primary = 36
	Success = 32
	Failed  = 31
	Cancel  = 34
)

func ColorWrapper(msg string, color int) string {
	if Logger.color {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, msg)
	} else {
		return msg
	}
}

func ColorWrapperInfo(color int, nodelist []string, msg string) {
	diviLine := ColorWrapper("--------------------", color)
	metaLine := ColorWrapper(fmt.Sprintf("%s  (%d)", utils.Merge(nodelist...), len(nodelist)), color)
	Infof("\n%s\n%s\n%s\n%s\n", diviLine, metaLine, diviLine, msg)
}
