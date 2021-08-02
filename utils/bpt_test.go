package utils

import (
	"testing"
)

func TestParseTree(t *testing.T) {
	bt := NewBpts("vm[0-1],pan0", 2)
	bt.Gen()
	nodelist := bt.ParseTree()
	t.Log(nodelist)
}
