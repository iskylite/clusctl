package service

import (
	"fmt"
	"myclush/utils"
)

// NodeSet format nodelist
type NodeSet struct {
	nodelist      []string
	isExpand      bool
	isMultiExpand bool
}

// NewNodeSet construct function
func NewNodeSet(nodes string, isExpand, isMultiExpand bool) *NodeSet {
	nodelist := utils.ExpNodes(nodes)
	return &NodeSet{
		nodelist:      nodelist,
		isExpand:      isExpand,
		isMultiExpand: isMultiExpand,
	}
}

func (n *NodeSet) getNodeList() []string {
	return n.nodelist
}

// Echo format nodelist
func (n *NodeSet) Echo() error {
	if n.isMultiExpand {
		for _, node := range n.getNodeList() {
			fmt.Println(node)
		}
		return nil
	}
	if n.isExpand {
		for _, node := range n.getNodeList() {
			fmt.Printf("%s ", node)
		}
		fmt.Println()
		return nil
	}
	nodes := utils.Merge(n.getNodeList()...)
	fmt.Println(nodes)
	return nil
}
