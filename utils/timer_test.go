package utils

import (
	"math"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	dateNanoUnix := NewTimerAfterSeconds(1)
	t.Log(time.Now().Local().String())
	t.Log(math.Ceil(time.Until(time.Unix(0, dateNanoUnix)).Seconds()))
	timer := GenTikerWithTimer(dateNanoUnix)
	<-timer
	t.Log(time.Now().Local().String())
}
