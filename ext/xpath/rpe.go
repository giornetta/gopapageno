package xpath

import (
	"fmt"
	"strings"
)

type rpe interface {
	udpe
}

type rpeBuilder interface {
	udpeBuilder
}

type rpeBuilderState int

const (
	expectRpeAxis rpeBuilderState = iota
	expectRpeUdpeTest
)

type rpeBuilderImpl struct {
	state            rpeBuilderState
	currentInnerTest *rpeInnerTestImpl
}

func newRpeBuilder() rpeBuilder {
	return new(rpeBuilderImpl)
}

func (rpeBuilder *rpeBuilderImpl) init() {
	return
}

func (rpeBuilder *rpeBuilderImpl) addUdpeTest(udpeTest udpeTest) (ok bool) {
	if rpeBuilder.state != expectRpeUdpeTest {
		ok = false
		return
	}
	defer func() {
		rpeBuilder.state = expectRpeAxis
	}()

	ok = true
	nextRpeInnerTest := &rpeInnerTestImpl{
		udpeTest:         udpeTest,
		nextRpeInnerTest: rpeBuilder.currentInnerTest,
	}
	rpeBuilder.currentInnerTest = nextRpeInnerTest
	return
}

func (rpeBuilder *rpeBuilderImpl) addAxis(axis axis) (ok bool) {
	if rpeBuilder.state != expectRpeAxis {
		ok = false
		return
	}
	defer func() {
		rpeBuilder.state = expectRpeUdpeTest
	}()

	ok = true
	if rpeBuilder.currentInnerTest == nil {
		return
	}
	if axis == ancestorOrSelf {
		rpeBuilder.currentInnerTest.behindAncestorAxis = true
	}
	return
}

func (rpeBuilder *rpeBuilderImpl) end() (result udpe) {
	if rpeBuilder.currentInnerTest != nil {
		rpeBuilder.currentInnerTest.isEntry = true
		result = &rpeImpl{
			entryTest: rpeBuilder.currentInnerTest,
		}
	}
	return
}

type rpeInnerTestImpl struct {
	isEntry            bool
	behindAncestorAxis bool
	udpeTest           udpeTest
	nextRpeInnerTest   *rpeInnerTestImpl
}

func (rpeInnerTest *rpeInnerTestImpl) matchWithReductionOf(n interface{}) (predicate predicate, next, newTest *rpeInnerTestImpl, hasNewTest, ok bool) {
	doesUdpeTestMatches := rpeInnerTest.udpeTest.test(n)
	if rpeInnerTest.isEntry {
		switch {
		case rpeInnerTest.behindAncestorAxis:
			ok = true
			next = rpeInnerTest
			if doesUdpeTestMatches {
				newTest = rpeInnerTest.nextRpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = rpeInnerTest.nextRpeInnerTest
		default:
			ok = false
		}
	} else {
		switch {
		case rpeInnerTest.behindAncestorAxis && rpeInnerTest.nextRpeInnerTest != nil && rpeInnerTest.nextRpeInnerTest.behindAncestorAxis:
			ok = true
			if doesUdpeTestMatches {
				next = rpeInnerTest.nextRpeInnerTest
			} else {
				next = rpeInnerTest
			}
		case rpeInnerTest.behindAncestorAxis:
			ok = true
			next = rpeInnerTest
			if doesUdpeTestMatches {
				newTest = rpeInnerTest.nextRpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = rpeInnerTest.nextRpeInnerTest
		default:
			ok = false
		}
	}

	if ok {
		predicate = rpeInnerTest.udpeTest.predicate()
	}
	return
}

func (rpeInnerTest *rpeInnerTestImpl) entry() bool {
	return rpeInnerTest.isEntry
}

func (rpeInnerTest *rpeInnerTestImpl) String() (result string) {
	if rpeInnerTest.behindAncestorAxis {
		result = `\\`
	} else if !rpeInnerTest.isEntry {
		result = `\`
	}
	result += fmt.Sprintf("%v", rpeInnerTest.udpeTest)
	return
}

//concrete rpe path pattern implementation
type rpePathPathPatternImpl struct {
	currentTest *rpeInnerTestImpl
}

func (rpePP *rpePathPathPatternImpl) isEmpty() bool {
	return rpePP.currentTest == nil
}

func (rpePP *rpePathPathPatternImpl) matchWithReductionOf(n interface{}, doUpdate bool) (predicate predicate, newPathPattern pathPattern, ok bool) {
	if rpePP.isEmpty() {
		panic(`rpe path pattern error: trying a match for an empty path pattern`)
	}

	predicate, next, newTest, hasNewTest, ok := rpePP.currentTest.matchWithReductionOf(n)
	if !ok {
		return
	}

	if !doUpdate {
		return
	}

	rpePP.currentTest = next
	if hasNewTest {
		newPathPattern = &rpePathPathPatternImpl{newTest}
	}
	return
}

func (rpePP *rpePathPathPatternImpl) String() (result string) {
	result = rpeStringifyUtil(rpePP.currentTest, rpePathPatternMode)

	if !strings.HasPrefix(result, `\\`) {
		result = strings.TrimPrefix(result, `\`)
	}
	return
}

//concrete rpe implementation
type rpeImpl struct {
	entryTest *rpeInnerTestImpl
}

func (rpe *rpeImpl) entryPoint() pathPattern {
	return &rpePathPathPatternImpl{rpe.entryTest}
}

func (rpe *rpeImpl) String() string {
	return rpeStringifyUtil(rpe.entryTest, rpeMode)
}

//utils
type rpeStringificationMode int

const (
	rpeMode rpeStringificationMode = iota
	rpePathPatternMode
)

func rpeStringifyUtil(currentTest *rpeInnerTestImpl, mode rpeStringificationMode) (result string) {
	for ct := currentTest; ct != nil; ct = ct.nextRpeInnerTest {
		result += ct.String()
	}
	if mode == rpePathPatternMode {
		result += `\ε`
	}
	return
}
