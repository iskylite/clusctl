package service

import (
	"testing"
)

func Test_hashNodesMap(t *testing.T) {
	type args struct {
		nodes string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{"hashNodesMap", args{nodes: "cn[0-200000]"}}, // 0.4s
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hashNodesMap(tt.args.nodes)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(got.Load("cn100"))
		})
	}
}
