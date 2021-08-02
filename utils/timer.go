package utils

import "time"

func LocalTime() string {
	return time.Now().Local().Format("2006-01-02T15:04:05")
}

func FormatTime(date int64) string {
	return time.Unix(0, date).Local().String()
}

func NewTimerAfterSeconds(afterSeconds int64) int64 {
	date := time.Now().Add(time.Second * time.Duration(afterSeconds))
	return date.UnixNano()
}

func GenTikerWithTimer(dateNanoUnix int64) <-chan time.Time {
	return time.After(time.Until(time.Unix(0, dateNanoUnix)))
}
