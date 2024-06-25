package gopapageno

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Token struct {
	Type       TokenType
	Precedence Precedence

	Value any

	Next      *Token
	Child     *Token
	LastChild *Token
}

func (t *Token) IsTerminal() bool {
	return t.Type.IsTerminal()
}

type TokenType uint16

const (
	TokenEmpty TokenType = 0
	TokenTerm  TokenType = 0x8000
)

func (t TokenType) IsTerminal() bool {
	return t >= 0x8000
}

func (t TokenType) Value() uint16 {
	return uint16(0x7FFF & t)
}

// Height computes the height of the AST rooted in `t`.
// It can be used as an evaluation metric for tree-balance, as left/right-skewed trees will have a bigger height compared to balanced trees.
func (root *Token) Height(ctx context.Context) (int, error) {
	// Helper struct to hold a token and its depth
	type TokenWithDepth struct {
		token *Token
		depth int
	}

	if root == nil {
		return 0, nil
	}

	var maxHeight int
	var wg sync.WaitGroup
	resultChan := make(chan int)

	// Launch the initial goroutine
	wg.Add(1)
	go bfs(root, 0, &wg, resultChan)

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Find the maximum height from the results
	for height := range resultChan {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}

		if height > maxHeight {
			maxHeight = height
		}
	}

	return maxHeight, nil
}
func bfs(node *Token, depth int, wg *sync.WaitGroup, resultChan chan int) {
	defer wg.Done()

	// Send the current depth to the result channel
	resultChan <- depth

	// Process the child node
	if node.Child != nil {
		wg.Add(1)
		go bfs(node.Child, depth+1, wg, resultChan)
	}

	// Process the next sibling node
	if node.Next != nil {
		wg.Add(1)
		go bfs(node.Next, depth, wg, resultChan)
	}
}

// Size returns the number of tokens in the AST rooted in `t`.
func (t *Token) Size() int {
	var rec func(t *Token, root bool) int

	rec = func(t *Token, root bool) int {
		if t == nil {
			return 0
		}

		if root {
			return 1 + rec(t.Child, false)
		} else {
			return 1 + rec(t.Child, false) + rec(t.Next, false)
		}
	}

	return rec(t, true)
}

// String returns a string representation of the AST rooted in `t`.
// This should be used rarely, as it doesn't print out a proper string representation of the token type.
// Gopapageno will generate a `SprintToken` function for your tokens.
func (t *Token) String() string {
	var sprintRec func(t *Token, sb *strings.Builder, indent string)

	sprintRec = func(t *Token, sb *strings.Builder, indent string) {
		if t == nil {
			return
		}

		sb.WriteString(indent)

		if t.Next == nil {
			sb.WriteString("└── ")
			indent += "    "
		} else {
			sb.WriteString("├── ")
			indent += "|   "
		}

		sb.WriteString(fmt.Sprintf("%d: %v\n", t.Type, t.Value))

		sprintRec(t.Child, sb, indent)
		sprintRec(t.Next, sb, indent[:len(indent)-4])
	}

	var sb strings.Builder

	sprintRec(t, &sb, "")

	return sb.String()
}
