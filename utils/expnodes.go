package utils

import (
	"github.com/iskylite/nodeset"
)

func ExpNodes(nodestr string) []string {
	if nodes, err := nodeset.Expand(nodestr); err != nil {
		return []string{}
	} else {
		return nodes
	}
}

func Merge(nodes ...string) string {
	if nodestr, err := nodeset.Merge(nodes...); err != nil {
		return ""
	} else {
		return nodestr
	}
}

func SplitNodesByWidth(nodes []string, width int32) [][]string {
	return nodeset.Split(nodes, int(width))
}
