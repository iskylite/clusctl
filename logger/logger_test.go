package logger

import "testing"

func TestDebug(t *testing.T) {
	SetLevel(DEBUG)
	// SetSilent()
	SetColor()
	Debug("This is a DEBUG log")
}

func TestDebugf(t *testing.T) {
	SetLevel(DEBUG)
	ResetSilent()
	Debugf("This is a %s log\n", "DEBUGF")
}

func TestInfo(t *testing.T) {
	SetLevel(INFO)
	ResetSilent()
	Info("This is a INFO log")
}

func TestWarning(t *testing.T) {
	SetLevel(WARNING)
	ResetSilent()
	Warning("This is a INFO log")
}

func TestError(t *testing.T) {
	SetLevel(ERROR)
	ResetSilent()
	Error("This is a INFO log")
}

func TestColorWrapper(t *testing.T) {
	t.Log(ColorWrapper("Primary: Hello, World", Primary))
	t.Log(ColorWrapper("Success: Hello, World", Success))
	t.Log(ColorWrapper("Failed: Hello, World", Failed))
	t.Log(ColorWrapper("Warn: Hello, World", Warn))
	t.Log(ColorWrapper("Cancel: Hello, World", Cancel))
}