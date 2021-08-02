package logger

import (
	"fmt"
	"log"
	"myclush/utils"
)

const (
	DEBUG = iota
	INFO
	WARNING
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

func newlogger(level int) *logger {
	defaultLogger := log.Default()
	defaultLogger.SetFlags(log.Lshortfile)
	return &logger{
		level:  level,
		silent: false,
		color:  false,
		log:    defaultLogger,
	}
}

func SetLevel(level int) {
	Logger.level = level
}

func SetSilent() {
	Logger.silent = true
}

func ResetSilent() {
	Logger.silent = false
}

func SetColor() {
	Logger.color = true
}

func ResetColor() {
	Logger.color = false
}

func GetLevel() int {
	return Logger.level
}

func Debug(a ...interface{}) {
	if Logger.level == DEBUG {
		if Logger.silent {
			fmt.Println(a...)
		} else {
			lp := "DEBUG => "
			if Logger.color {
				lp = "\x1b[32mDEBUG => \x1b[0m"
			}
			v := append([]interface{}{utils.LocalTime(), lp}, a...)
			Logger.log.Output(2, fmt.Sprintln(v...))
		}
	}
}

func Debugf(format string, v ...interface{}) {
	if Logger.level == DEBUG {
		if Logger.silent {
			fmt.Printf(format, v...)
		} else {
			lp := "DEBUG =>"
			if Logger.color {
				lp = "\x1b[32mDEBUG =>\x1b[0m"
			}
			levelFormat := fmt.Sprintf("%s %s %s", utils.LocalTime(), lp, format)
			Logger.log.Output(2, fmt.Sprintf(levelFormat, v...))
		}
	}
}
func Info(a ...interface{}) {
	if Logger.level <= INFO {
		if Logger.silent {
			fmt.Println(a...)
		} else {
			lp := "INFO => "
			if Logger.color {
				lp = "\x1b[36mINFO => \x1b[0m"
			}
			v := append([]interface{}{utils.LocalTime(), lp}, a...)
			Logger.log.Output(2, fmt.Sprintln(v...))
		}
	}
}

func Infof(format string, v ...interface{}) {
	if Logger.level <= INFO {
		if Logger.silent {
			fmt.Printf(format, v...)
		} else {
			lp := "INFO =>"
			if Logger.color {
				lp = "\x1b[36mINFO =>\x1b[0m"
			}
			levelFormat := fmt.Sprintf("%s %s %s", utils.LocalTime(), lp, format)
			Logger.log.Output(2, fmt.Sprintf(levelFormat, v...))
		}
	}
}
func Warning(a ...interface{}) {
	if Logger.level <= WARNING {
		if Logger.silent {
			fmt.Println(a...)
		} else {
			lp := "WARN => "
			if Logger.color {
				lp = "\x1b[33mWARN => \x1b[0m"
			}
			v := append([]interface{}{utils.LocalTime(), lp}, a...)
			Logger.log.Output(2, fmt.Sprintln(v...))
		}
	}
}

func Warningf(format string, v ...interface{}) {
	if Logger.level <= WARNING {
		if Logger.silent {
			fmt.Printf(format, v...)
		} else {
			lp := "WARN =>"
			if Logger.color {
				lp = "\x1b[33mWARN =>\x1b[0m"
			}
			levelFormat := fmt.Sprintf("%s %s %s", utils.LocalTime(), lp, format)
			Logger.log.Output(2, fmt.Sprintf(levelFormat, v...))
		}
	}
}
func Error(a ...interface{}) {
	if Logger.level <= ERROR {
		if Logger.silent {
			fmt.Println(a...)
		} else {
			lp := "ERROR => "
			if Logger.color {
				lp = "\x1b[31mERROR => \x1b[0m"
			}
			v := append([]interface{}{utils.LocalTime(), lp}, a...)
			Logger.log.Output(2, fmt.Sprintln(v...))
		}
	}
}

func Errorf(format string, v ...interface{}) {
	if Logger.level <= ERROR {
		if Logger.silent {
			fmt.Printf(format, v...)
		} else {
			lp := "ERROR => "
			if Logger.color {
				lp = "\x1b[31mERROR => \x1b[0m"
			}
			levelFormat := fmt.Sprintf("%s %s%s", utils.LocalTime(), lp, format)
			Logger.log.Output(2, fmt.Sprintf(levelFormat, v...))
		}
	}
}

const (
	Primary = 36
	Success = 32
	Failed  = 31
	Warn    = 33
	Cancel  = 34
)

func ColorWrapper(msg string, color int) string {
	if Logger.color {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, msg)
	} else {
		return msg
	}
}
