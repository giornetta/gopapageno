package xpath

import "fmt"

// NonTerminal represents a unique non terminal inside the syntax tree representing
// the XML document
type NonTerminal interface {
	SetExecutionTable(execTab executionTable) NonTerminal
	SetNode(n interface{}) NonTerminal
	Children() []NonTerminal
	SetDirectChildAndInheritItsChildren(NonTerminal) NonTerminal
	ExecutionTable() executionTable
	Node() interface{}
	Position() Position
}

func NewNonTerminal() NonTerminal {
	return &nonTerminalImpl{}
}

type nonTerminalImpl struct {
	n       interface{}
	ch      []NonTerminal
	execTab executionTable
}

func (nt *nonTerminalImpl) String() string {
	if nt == nil {
		return "-"
	}
	return fmt.Sprintf("E(%p)", nt)
}

func (nt *nonTerminalImpl) SetExecutionTable(executionTable executionTable) NonTerminal {
	nt.execTab = executionTable
	return nt
}

func (nt *nonTerminalImpl) ExecutionTable() executionTable {
	return nt.execTab
}

func (nt *nonTerminalImpl) SetNode(n interface{}) NonTerminal {
	nt.n = n
	return nt
}

func (nt *nonTerminalImpl) Node() interface{} {
	return nt.n
}

func (nt *nonTerminalImpl) SetDirectChildAndInheritItsChildren(child NonTerminal) NonTerminal {
	nt.ch = append(child.Children(), child)
	return nt
}

func (nt *nonTerminalImpl) Children() []NonTerminal {
	return nt.ch
}

func (nt *nonTerminalImpl) Position() Position {
	if element, isElement := nt.n.(*Element); isElement {
		return element.position()
	}

	if text, isText := nt.n.(*Text); isText {
		return text.position()
	}

	return nil
}

// Position represents the Position of some information inside a document
// in terms of number of characters from the beginning of the document.
type Position interface {
	Extremes() (start, end int)
	Start() int
	End() int
}

type position struct {
	start, end int
}

func newPosition(start, end int) *position {
	return &position{start, end}
}

func (p *position) String() string {
	return fmt.Sprintf("(%d , %d)", p.start, p.end)
}

func (p *position) Start() int {
	return p.start
}

func (p *position) End() int {
	return p.end
}

func (p *position) Extremes() (start, end int) {
	start = p.start
	end = p.end
	return
}

type semanticValue struct {
	id string

	startPos int
	endPos   int
}

type OpenTagSemanticValue struct {
	semanticValue

	attributes []*Attribute
}

func NewOpenTagSemanticValue(id string, attributes []*Attribute, startPos int, endPos int) *OpenTagSemanticValue {
	return &OpenTagSemanticValue{
		semanticValue: semanticValue{
			id:       id,
			startPos: startPos,
			endPos:   endPos,
		},
		attributes: attributes,
	}
}

type CloseTagSemanticValue struct {
	semanticValue
}

func NewCloseTagSemanticValue(id string, startPos int, endPos int) *CloseTagSemanticValue {
	return &CloseTagSemanticValue{
		semanticValue{
			id:       id,
			startPos: startPos,
			endPos:   endPos,
		},
	}
}

type OpenCloseTagSemanticValue struct {
	OpenTagSemanticValue
}

func NewOpenCloseTagSemanticValue(id string, attributes []*Attribute, startPos int, endPos int) *OpenCloseTagSemanticValue {
	return &OpenCloseTagSemanticValue{
		OpenTagSemanticValue{
			semanticValue: semanticValue{
				id:       id,
				startPos: startPos,
				endPos:   endPos,
			},
			attributes: attributes,
		},
	}
}

type TextSemanticValue struct {
	data string

	startPos int
	endPos   int
}

func NewTextSemanticValue(data string, startPos int, endPos int) *TextSemanticValue {
	return &TextSemanticValue{
		data:     data,
		startPos: startPos,
		endPos:   endPos,
	}
}
