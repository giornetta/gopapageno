package xpath

import (
	"fmt"
	"strings"
)

type axis int

const (
	child axis = iota
	descendantOrSelf
	parent
	ancestorOrSelf
)

type udpeType int

const (
	FPE udpeType = iota
	RPE
)

func (ut udpeType) String() string {
	switch ut {
	case FPE:
		return "FPE"
	case RPE:
		return "RPE"
	default:
		return "NaUdpe"
	}
}

func (a axis) String() (s string) {
	switch a {
	case child:
		s = "child"
	case descendantOrSelf:
		s = "descendant-or-self"
	case parent:
		s = "parent"
	case ancestorOrSelf:
		s = "ancestor-or-self"
	}
	return
}

type udpeTest interface {
	test(tested interface{}) bool
	predicate() predicate
}

type pathPattern interface {
	matchWithReductionOf(n interface{}, doUpdate bool) (predicate predicate, newPathPattern pathPattern, ok bool)
	isEmpty() bool
	String() string
}

type udpe interface {
	entryPoint() pathPattern
	String() string
}

type udpeBuilder interface {
	init()
	addUdpeTest(udpeTest udpeTest) (ok bool)
	addAxis(axis axis) (ok bool)
	end() udpe
}

type elementTest struct {
	wildCard bool
	name     string
	attr     *Attribute
	pred     predicate
}

func newElementTest(name string, attribute *Attribute, predicate predicate) *elementTest {
	return &elementTest{
		name:     name,
		wildCard: name == "*",
		attr:     attribute,
		pred:     predicate,
	}
}

func (et *elementTest) predicate() predicate {
	if et.pred != nil {
		return et.pred.copy()
	}
	return nil
}

func (et *elementTest) test(tested interface{}) bool {
	element, isElement := tested.(*Element)

	if !isElement {
		return false
	}

	elementName := element.name
	if !et.wildCard && elementName != et.name {
		return false
	}

	if et.attr != nil {
		hasAttribute := false
		for _, attribute := range element.attributes {
			if *attribute == *et.attr {
				hasAttribute = true
				break
			}
		}
		return hasAttribute
	}

	return true
}

func (et *elementTest) String() string {
	result := []string{et.name}
	if et.pred != nil {
		result = append(result, "[p]")
	}
	if et.attr != nil {
		result = append(result, fmt.Sprintf("/@%v", et.attr))
	}
	return strings.Join(result, "")
}

type textTest struct {
	data string
}

func newTextTest(text string) *textTest {
	return &textTest{text}
}

func (tt *textTest) predicate() predicate {
	return nil
}

func (tt *textTest) test(node interface{}) bool {
	text, isText := node.(*Text)

	if !isText {
		return false
	}
	return tt.data == "" || tt.data == text.data
}

func (tt *textTest) String() string {
	return fmt.Sprintf("Text(%q)", tt.data)
}
