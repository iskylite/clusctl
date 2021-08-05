package utils

import (
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

func getPrefix(node string) string {
	re := regexp.MustCompile(`^[a-zA-Z]*`)
	return string(re.Find([]byte(node)))
}

func stringToInt(str string) int {
	ans := 0
	for _, char := range str {
		ans = ans*10 + int(char) - '0'
	}
	return ans
}

func divNum(numSection string) []int {
	str := strings.Split(numSection, "-")
	//fmt.Println(numSection, " ", str)
	begin := stringToInt(str[0])
	end := begin
	if len(str) == 2 {
		end = stringToInt(str[1])
	}
	//fmt.Printf("len(str) = %d, begin = %d, end = %d\n", len(str), begin, end)
	Nums := make([]int, 0)
	for i := begin; i <= end; i++ {
		Nums = append(Nums, i)
	}
	return Nums
}

// 模式1: cn[0-1,2-3,cn4] -> cn0,cn1,cn2,cn3,cn4
// 模式2: cn0 -> cn0
// 模式3: cn -> cn
func divNode(nodeList string, nodeChan chan string) {
	prefix := getPrefix(nodeList)
	//fmt.Printf("nodeList = %v, prefix = %s\n", node, prefix)
	if len(nodeList) > len(prefix) {
		if nodeList[len(prefix)] == '[' && nodeList[len(nodeList)-1] == ']' { //模式1
			begin := len(prefix) + 1
			Nums := make([]int, 0)
			for idx, char := range nodeList {
				if char == ',' {
					//fmt.Printf("begin = %d, idx = %d\n", begin, idx)
					Nums = append(Nums, divNum(nodeList[begin:idx])...)
					begin = idx + 1
				}
			}
			if nodeList[len(nodeList)-2] != ',' {
				Nums = append(Nums, divNum(nodeList[begin:len(nodeList)-1])...)
			}

			for _, num := range Nums {
				nodeChan <- fmt.Sprintf("%s%d", prefix, num)
			}
		} else { //模式2
			nodeChan <- nodeList
		}
	} else { //模式3
		nodeChan <- nodeList
	}
}

// cn0,cn[2-3],cn[4-5,6-7] -> {cn0,cn2,cn3,cn4,cn5,cn6,cn7}
func expNodes(nodeListStr string, nodeChan chan string) {
	nodeSection := splitNodeList(nodeListStr)
	for _, nodeList := range nodeSection {
		divNode(nodeList, nodeChan)
	}
}

// add node to nodeChan
func AddNode(nodeList string, nodeChan chan string) {
	defer close(nodeChan)
	expNodes(nodeList, nodeChan)
}

func splitNodeList(nodeListStr string) []string {
	flag := true
	begin := 0
	nodeList := make([]string, 0)
	for idx, char := range nodeListStr {
		if char == '[' {
			flag = false
		} else if char == ']' {
			flag = true
		} else if char == ',' {
			if flag {
				nodeList = append(nodeList, nodeListStr[begin:idx])
				begin = idx + 1
			}
		}
	}
	if nodeListStr[len(nodeListStr)-1] != ',' {
		nodeList = append(nodeList, nodeListStr[begin:])
	}
	return nodeList
}

func ExpNodes(nodelist string) []string {
	nodes := make([]string, 0)
	if nodelist == "" {
		return nodes
	}
	nodesChan := make(chan string, runtime.NumCPU())
	go AddNode(nodelist, nodesChan)
	for node := range nodesChan {
		nodes = append(nodes, node)
	}
	sort.Sort(mySort(nodes))
	return nodes
}

type mySort []string

func (m mySort) Len() int {
	return len(m)
}

func (m mySort) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m mySort) Parse(i int) (string, int) {
	if m.Len() < 1 {
		return "", -1
	}
	re := regexp.MustCompile(`^([a-zA-Z]+)([0-9]+)$`)
	matches := re.FindAllStringSubmatch(m[i], -1)
	if len(matches) == 0 {
		return "", -1
	}
	index, _ := strconv.Atoi(matches[0][2])
	return matches[0][1], index
}

func (m mySort) Less(i, j int) bool {
	fpre, findex := m.Parse(i)
	// fmt.Printf("%s%d\n", fpre, findex)
	if findex == -1 {
		return false
	}
	spre, sindex := m.Parse(j)
	// fmt.Printf("%s%d\n", spre, sindex)
	if sindex == -1 {
		return true
	}
	if fpre == spre {
		return findex < sindex
	}
	if fpre > spre {
		return false
	}
	return true
}

func ConvertNodelist(nodelist []string) string {
	if len(nodelist) < 1 {
		return ""
	}
	sort.Sort(mySort(nodelist))
	prefix := ""
	start := ""
	end := ""
	nodes := ""
	for _, node := range nodelist {
		lprefix := getPrefix(node)
		lcntStr := strings.TrimPrefix(node, lprefix)
		if prefix == "" {
			prefix = lprefix
			start, end = lcntStr, lcntStr
			nodes = fmt.Sprintf("%s[%s", prefix, start)
			continue
		}
		if prefix == lprefix {
			lcnt, _ := strconv.Atoi(lcntStr)
			lend, _ := strconv.Atoi(end)
			if lcnt-lend == 1 {
				end = lcntStr
			} else {
				if start != end {
					nodes = fmt.Sprintf("%s-%s,%s", nodes, end, lcntStr)
				} else {
					nodes = fmt.Sprintf("%s,%s", nodes, lcntStr)
				}
				start, end = lcntStr, lcntStr
			}
		} else {
			prefix = lprefix
			if start != end {
				nodes = fmt.Sprintf("%s-%s],%s[%s", nodes, end, prefix, lcntStr)
			} else {
				suffixStr := fmt.Sprintf("[%s", start)
				if strings.HasSuffix(nodes, suffixStr) {
					nodes = fmt.Sprintf("%s%s,%s[%s", strings.TrimSuffix(nodes, suffixStr), start, prefix, lcntStr)
				} else {
					nodes = fmt.Sprintf("%s],%s[%s", nodes, prefix, lcntStr)
				}
			}
			start, end = lcntStr, lcntStr
		}
	}
	if start == end {
		suffixStr := fmt.Sprintf("[%s", start)
		if strings.HasSuffix(nodes, suffixStr) {
			nodes = fmt.Sprintf("%s%s", strings.TrimSuffix(nodes, suffixStr), start)
		} else {
			nodes = fmt.Sprintf("%s]", nodes)
		}
	} else {
		nodes = fmt.Sprintf("%s-%s]", nodes, end)
	}
	return nodes
}

func SplitNodesByWidth(nodelist []string, width int32) [][]string {
	splitNodelist := make([][]string, 0)
	if len(nodelist) == 0 {
		return splitNodelist
	}
	w := int(width)
	nodesNum := len(nodelist)
	if nodesNum <= w {
		for _, node := range nodelist {
			splitNodelist = append(splitNodelist, []string{node})
		}
	} else {
		splitNums := nodesNum / w
		leaveNums := nodesNum % w
		i := 0
		for i = 0; i < w; i++ {
			s := i * splitNums
			e := (i + 1) * splitNums
			if i < leaveNums {
				if i == 0 {
					e++
				} else {
					s += i
					e += i + 1
				}
			} else {
				s += leaveNums
				e += leaveNums
			}
			splitNodelist = append(splitNodelist, nodelist[s:e])
		}
	}
	return splitNodelist
}
