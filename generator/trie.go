package generator

import (
	"fmt"
	"github.com/giornetta/gopapageno"
)

// triePtrNode has a TokenType as key and a pointer to a treeNode as value.
type triePtrNode struct {
	Key gopapageno.TokenType
	Ptr *trieNode
}

// trieNode is a node of a trie.
// It has pointers to other trieNodes and may have (if it's the terminal node of a rhs) a value (the corresponding lhs)
// as well as the number of the rule (needed for the semantic function).
type trieNode struct {
	HasValue bool
	Value    gopapageno.TokenType
	RuleNum  int
	Branches []triePtrNode
}

// Get obtains the node that is assigned to a certain key.
func (n *trieNode) Get(key gopapageno.TokenType) (*trieNode, bool) {
	branches := n.Branches
	low := 0
	high := len(branches) - 1

	for low <= high {
		curPos := low + (high-low)/2
		curKey := branches[curPos].Key

		if key < curKey {
			high = curPos - 1
		} else if key > curKey {
			low = curPos + 1
		} else {
			return branches[curPos].Ptr, true
		}
	}

	return nil, false
}

// Find traverses a trie using the elements in rhs as keys, and returns the last node.
func (n *trieNode) Find(rhs []gopapageno.TokenType) (*trieNode, bool) {
	curNode := n
	for _, token := range rhs {
		nextNode, ok := curNode.Get(token)
		if !ok {
			return nil, false
		}
		curNode = nextNode
	}

	return curNode, true
}

/*
newTrie creates a trie from a set of rules and returns it.
The rhs of the rules must be sorted.
*/
func newTrie(rules []rule, nonterminals *set[string], terminals *set[string]) (*trieNode, error) {
	root := &trieNode{false, 0, 0, make([]triePtrNode, 0)}

	nonterminalSlice := nonterminals.Slice()
	if err := moveToFront(nonterminalSlice, "_EMPTY"); err != nil {
		return nil, err
	}

	terminalSlice := terminals.Slice()
	if err := moveToFront(terminalSlice, "_TERM"); err != nil {
		return nil, err
	}

	for i, rule := range rules {
		curNode := root
		for j, strToken := range rule.RHS {
			// Find the current token in the nodes' children.
			token := tokenValue(strToken, nonterminalSlice, terminalSlice)
			nextNode, ok := curNode.Get(token)
			// If not found, create a new empty node and append it as a child to the current one.
			if !ok {
				nextNode = &trieNode{false, 0, 0, make([]triePtrNode, 0)}
				curNode.Branches = append(curNode.Branches, triePtrNode{token, nextNode})
			}
			curNode = nextNode

			// If we're at the last token of the rule, assign the current LHS to the last node created.
			if j == len(rule.RHS)-1 {
				curNode.HasValue = true
				curNode.Value = tokenValue(rule.LHS, nonterminalSlice, terminalSlice)
				curNode.RuleNum = i
			}
		}
	}

	return root, nil
}

func (n *trieNode) compressR(newTrie *[]gopapageno.TokenType, curpos *uint16, nonterminals *set[string], terminals *set[string]) {
	//Append the value of this node if it has one, and the rule number
	if n.HasValue {
		*newTrie = append(*newTrie, n.Value)
		*newTrie = append(*newTrie, gopapageno.TokenType(n.RuleNum))
	} else {
		*newTrie = append(*newTrie, gopapageno.TokenEmpty)
		*newTrie = append(*newTrie, 0)
	}
	*curpos += 2

	//Append the number of indices of this node
	*newTrie = append(*newTrie, gopapageno.TokenType(len(n.Branches)))
	*curpos++

	startPos := *curpos
	for i := 0; i < len(n.Branches); i++ {
		*newTrie = append(*newTrie, n.Branches[i].Key)
		//The offset will be updated later
		*newTrie = append(*newTrie, 0)
		*curpos += 2
	}

	for i := 0; i < len(n.Branches); i++ {
		//Update the offset
		(*newTrie)[uint16(startPos)+1+uint16(i)*2] = gopapageno.TokenType(*curpos)
		//Call compress on the pointed node
		n.Branches[i].Ptr.compressR(newTrie, curpos, nonterminals, terminals)
	}
}

func (n *trieNode) Compress(nonterminals *set[string], terminals *set[string]) []gopapageno.TokenType {
	compressedTrie := make([]gopapageno.TokenType, 0)
	curPos := uint16(0)

	n.compressR(&compressedTrie, &curPos, nonterminals, terminals)

	return compressedTrie
}

func tokenValue(token string, nonterminals []string, terminals []string) gopapageno.TokenType {
	for i, t := range nonterminals {
		if token == t {
			// If it's _EMPTY, return TokenEmpty (it should be at i == 0)
			// Otherwise start counting.
			return gopapageno.TokenEmpty + gopapageno.TokenType(i)
		}
	}

	for i, t := range terminals {
		if token == t {
			// If it's _TERM, return TokenTerm (it should be at i == 0)
			// Otherwise start counting.
			return gopapageno.TokenTerm + gopapageno.TokenType(i)
		}
	}

	return gopapageno.TokenEmpty
}

func (n *trieNode) printR(curDepth int) {
	if n.HasValue {
		fmt.Println(" ->", n.Value)
	} else {
		fmt.Println()
	}
	for _, nodePtr := range n.Branches {
		for i := 0; i < curDepth; i++ {
			fmt.Print("  ")
		}

		fmt.Print(nodePtr.Key)
		nodePtr.Ptr.printR(curDepth + 1)
	}
}

// Println prints a representation of the trie.
func (n *trieNode) Println() {
	n.printR(0)
	fmt.Println()
}
