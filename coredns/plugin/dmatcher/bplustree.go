package dmatcher

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shivamMg/ppds/tree"
)

type BPTree struct {
	Root     *Node
	FilePath string
}

type Node struct {
	Childs []*Node
	Value  rune
	HasEnd bool
}

func NewTree(filePath string) *BPTree {
	ins := &BPTree{
		Root:     &Node{Value: '.'},
		FilePath: filePath,
	}
	ins.Load()
	return ins
}

func (t *BPTree) Insert(value []rune) {
	t.Root.insert(value)
	t.Root.merge()

}
func (t *BPTree) Contain(value []rune) bool {
	return t.Root.contain(value)
}

func (t *BPTree) PrintAll() {
	t.Root.print(make([]rune, 0))
}
func (t *BPTree) GetAll(domain string) []string {
	list := make([]string, 0)
	t.Root.tolist(make([]rune, 0), &list, reverseString(domain))
	return list
}

func (n *Node) insert(value []rune) {
	//end leaf? quit
	if len(value) == 1 {
		n.HasEnd = true
		return
	}
	//has next token
	next := value[1:]

	//find and insert to exist node
	for _, node := range n.Childs {
		if node.Value == next[0] {
			node.insert(next)
			return
		}
	}
	//add new child node and insert inner
	nNode := &Node{Value: next[0]}
	nNode.insert(next)
	n.Childs = append(n.Childs, nNode)
}
func (n *Node) merge() {
	asterisk := false
	for _, node := range n.Childs {
		if node.Value == '*' {
			asterisk = true
			break
		}
	}
	if asterisk {
		repNode := &Node{Value: '*'}
		for _, child := range n.Childs {
			runes := child.extract()
			for _, word := range runes {
				repNode.insert(word)
			}
		}

		n.Childs = make([]*Node, 0)
		n.Childs = append(n.Childs, repNode)
	}
	for _, node := range n.Childs {
		node.merge()
	}
}

func (n *Node) extract() [][]rune {
	if len(n.Childs) == 0 {
		return [][]rune{{n.Value}}
	}
	res := make([][]rune, 0)
	for _, node := range n.Childs {
		for _, ss := range node.extract() {
			sub := make([]rune, 0)
			sub = append(sub, n.Value)
			sub = append(sub, ss...)
			res = append(res, sub)
		}
	}
	if n.HasEnd {
		res = append(res, []rune{n.Value})
	}
	return res
}
func (n *Node) contain(value []rune) bool {
	if value == nil {
		return false
	}
	if len(value) == 0 {
		return false
	}
	if value[0] == n.Value || n.Value == '*' {
		if len(value) > 1 {
			next := value[1:]
			for _, node := range n.Childs {
				res := node.contain(next)
				if res {
					return true
				}
			}
		}
		return n.HasEnd
	}
	return false
}

func (n *Node) print(value []rune) {
	value = append(value, n.Value)
	if len(n.Childs) == 0 && n.HasEnd {
		fmt.Printf("%s\n", reverseString(string(value)))
	} else {
		for _, node := range n.Childs {
			node.print(value)
		}
	}
}
func (n *Node) tolist(value []rune, list *[]string, domain string) {
	value = append(value, n.Value)
	if len(n.Childs) == 0 && n.HasEnd {
		if strings.HasPrefix(string(value), domain) {
			*list = append(*list, reverseString(string(value)))
		}
		//fmt.Println("node end: " + reverseString(string(value)))
	} else {
		for _, node := range n.Childs {
			node.tolist(value, list, domain)
		}
	}
}

// pretty print
// github.com/shivamMg/ppds/tree
func (n *Node) Data() interface{} {
	return string(n.Value)
}
func (n *Node) Children() (c []tree.Node) {
	for _, child := range n.Childs {
		c = append(c, tree.Node(child))
	}
	return
}

func reverseString(str string) string {
	byte_str := []rune(str)
	for i, j := 0, len(byte_str)-1; i < j; i, j = i+1, j-1 {
		byte_str[i], byte_str[j] = byte_str[j], byte_str[i]
	}
	return string(byte_str)
}

// interface implement
func (t *BPTree) AddDomain(val string) error {
	if strings.LastIndex(val, ".") != len(val) {
		val = val + "."
	}
	t.Insert([]rune(reverseString(val)))
	return nil
}
func (t *BPTree) DelDomain(val string) error {
	return errors.New("not implemented")
}
func (t *BPTree) ContainDomain(val string) (bool, error) {
	return t.Contain([]rune(reverseString(val))), nil
}
func (t *BPTree) GetDomainList(q string) ([]string, error) {
	return t.GetAll(q), nil
}
func (t *BPTree) Load() {
	file, err := os.Open(t.FilePath)
	if err != nil {
		log.Println("failed to load: " + err.Error())
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	count := 0
	for scanner.Scan() {
		//fmt.Printf("readed line %v %v\n", count, scanner.Text())
		count++
		t.AddDomain(scanner.Text())
	}
}
func (t *BPTree) Save() {

}
