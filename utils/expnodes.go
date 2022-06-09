package utils

import (
	"bufio"
	"os"
	"strings"

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

// ExpNodesFromFile read node or ip from hostfile
func ExpNodesFromFile(file string) ([]string, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	NodeList := make([]string, 0)
	buffer := bufio.NewScanner(fp)
	for buffer.Scan() {
		line := buffer.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// filter comment
			continue
		}
		NodeList = append(NodeList, line)
	}
	return NodeList, nil
}
