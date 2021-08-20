package logger

import (
	"testing"
)

func TestDebug(t *testing.T) {
	SetLevel(DEBUG)
	// SetSilent()
	Debug("This is a DEBUG log")
}

func TestDebugf(t *testing.T) {
	SetLevel(DEBUG)
	Debugf("This is a %s log\n", "DEBUGF")
}

func TestInfo(t *testing.T) {
	SetLevel(INFO)
	Info("This is a INFO log")
}

func TestError(t *testing.T) {
	SetLevel(ERROR)
	Error("This is a INFO log")
}

func TestColorWrapper(t *testing.T) {
	t.Log(ColorWrapper("Primary: Hello, World", Primary))
	t.Log(ColorWrapper("Success: Hello, World", Success))
	t.Log(ColorWrapper("Failed: Hello, World", Failed))
}
