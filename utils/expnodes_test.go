package utils

import (
	"sort"
	"testing"
)

func TestConvertNodelist(t *testing.T) {
	nodelist := []string{"pan0", "cn0", "cn2", "cn3", "cn4", "cn10", "vm0", "vm1"}
	// nodelist := []string{"cn9", "vm0", "vm1", "pan0"}
	t.Log(ConvertNodelist(nodelist))
}

func TestSplitNodesByWidth(t *testing.T) {
	nodes := ExpNodes("vm[0-85]")
	var width int32 = 7
	for _, node := range SplitNodesByWidth(nodes, width) {
		t.Log(node)
		t.Log(ConvertNodelist(node))
	}
}


func TestMySort(t *testing.T) {
	nodelist := []string{"pan0", "cn0", "vn1", "cn2", "cn3", "pan1", "cn4", "cn10", "vm0", "vm1"}
	t.Log(ConvertNodelist(nodelist))
	sort.Sort(mySort(nodelist))
	t.Log(nodelist)
	t.Log(ConvertNodelist(nodelist))
}

func TestExpNodes(t *testing.T) {
	t.Log(ExpNodes("cn[0-3,5-9,12-71,74-81,84-125,128-243]"))
}