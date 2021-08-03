package utils

import "testing"

func TestConvertSize(t *testing.T) {
	for _, size := range []string{"51200", "512K", "1m"} {
		s, err := ConvertSize(size)
		if err != nil {
			t.Error(err)
			continue
		}
		t.Logf("%s => %d\n", size, s)
	}
}
