package utils

import (
	pb "myclush/pb"
	"runtime"
)

type Bpts struct {
	nodes    string
	tree     *pb.Node
	nodelist []string
	width    int32
}

func NewBpts(nodes string, width int32) *Bpts {
	return &Bpts{
		nodes:    nodes,
		tree:     &pb.Node{},
		nodelist: make([]string, 0),
		width:    width,
	}
}

func NewBptsWithTree(tree *pb.Node) *Bpts {
	return &Bpts{
		tree:     tree,
		nodelist: make([]string, 0),
	}
}

func (b *Bpts) SetWidth(width int32) {
	b.width = width
}

func (b *Bpts) GetWidth() int32 {
	return b.width
}

func (b *Bpts) GetTree() *pb.Node {
	return b.tree
}

func (b *Bpts) Gen() {
	nodesChan := make(chan string, runtime.NumCPU())
	go AddNode(b.nodes, nodesChan)
	for node := range nodesChan {
		b.nodelist = append(b.nodelist, node)
	}
	// fmt.Println(b.nodelist)
	b.tree = insertNode(b.nodelist, 1, b.width)
}

func insertNode(nodelist []string, index int, width int32) *pb.Node {
	node := newNode(nodelist[index-1], width)
	for i := 0; i < int(width); i++ {
		nindex := index*2 + i
		if nindex > len(nodelist) {
			break
		}
		node.Nodes = append(node.Nodes, insertNode(nodelist, nindex, width))
	}
	return node
}

func (b *Bpts) ParseTree() []string {
	// return ReverseSlice(parseNodeTree(b.tree, nodelist))
	return parseNodeTree(b.tree)
}

func parseNodeTree(pn *pb.Node) []string {
	nodelist := make([]string, 0)
	if pn == nil {
		return nodelist
	}
	nodelist = append(nodelist, pn.Value)
	for _, n := range pn.Nodes {
		nodelist = append(nodelist, parseNodeTree(n)...)
	}
	return nodelist
}

func ReverseSlice(nodelist []string) []string {
	for i := len(nodelist)/2 - 1; i >= 0; i-- {
		opp := len(nodelist) - 1 - i
		nodelist[i], nodelist[opp] = nodelist[opp], nodelist[i]
	}
	return nodelist
}

func newNode(node string, width int32) *pb.Node {
	// fmt.Printf("Add Node [%s]\n", node)
	return &pb.Node{
		Value: node,
		Nodes: make([]*pb.Node, 0),
		Width: width,
	}
}
