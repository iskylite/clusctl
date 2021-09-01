package utils

import (
	"testing"
)

func TestMerge(t *testing.T) {
	t.Run("localhost_T", func(t *testing.T) {
		got := Merge([]string{"cn0", "cn3", "localhost"}...)
		t.Log(got)
	})
}
