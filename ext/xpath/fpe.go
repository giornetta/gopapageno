package xpath

import (
	"fmt"
	"strings"
)

type fpe interface {
	udpe
}

type fpeBuilder interface {
	udpeBuilder
}

type fpeBuilderState int

const (
	expectFpeAxis fpeBuilderState = iota
	expectFpeUdpeTest
)

//concrete implementation of a fpe builder
type fpeBuilderImpl struct {
	state                 fpeBuilderState
	precedentFpeInnerTest *fpeInnerTestImpl
}

func newFpeBuilder() fpeBuilder {
	return new(fpeBuilderImpl)
}

func (fpeBuilder *fpeBuilderImpl) init() {
	return
}

func (fpeBuilder *fpeBuilderImpl) addUdpeTest(udpeTest udpeTest) (ok bool) {
	if fpeBuilder.state != expectFpeUdpeTest {
		ok = false
		return
	}
	defer func() {
		fpeBuilder.state = expectFpeAxis
	}()

	ok = true
	nextFpeInnerTest := &fpeInnerTestImpl{
		udpeTest:              udpeTest,
		precedingFpeInnerTest: fpeBuilder.precedentFpeInnerTest,
	}

	fpeBuilder.precedentFpeInnerTest = nextFpeInnerTest
	return
}

func (fpeBuilder *fpeBuilderImpl) addAxis(axis axis) (ok bool) {
	if fpeBuilder.state != expectFpeAxis {
		ok = false
		return
	}
	defer func() {
		fpeBuilder.state = expectFpeUdpeTest
	}()

	ok = true
	if fpeBuilder.precedentFpeInnerTest == nil {
		return
	}

	if axis == descendantOrSelf {
		fpeBuilder.precedentFpeInnerTest.behindDescendantAxis = true
	}
	return
}

func (fpeBuilder *fpeBuilderImpl) end() (result udpe) {
	if fpeBuilder.precedentFpeInnerTest != nil {
		fpeBuilder.precedentFpeInnerTest.isEntry = true
		result = &fpeImpl{
			entryTest: fpeBuilder.precedentFpeInnerTest,
		}
	}
	return
}

type fpeInnerTestImpl struct {
	isEntry               bool
	behindDescendantAxis  bool
	udpeTest              udpeTest
	precedingFpeInnerTest *fpeInnerTestImpl
}

func (fpeInnerTest *fpeInnerTestImpl) matchWithReductionOf(n interface{}) (predicate predicate, next, newTest *fpeInnerTestImpl, hasNewTest, ok bool) {
	doesUdpeTestMatches := fpeInnerTest.udpeTest.test(n)
	if fpeInnerTest.isEntry {
		switch {
		case fpeInnerTest.behindDescendantAxis:
			ok = true
			next = fpeInnerTest
			if doesUdpeTestMatches {
				newTest = fpeInnerTest.precedingFpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = fpeInnerTest.precedingFpeInnerTest
		default:
			ok = false
		}
	} else {
		switch {
		case fpeInnerTest.behindDescendantAxis && fpeInnerTest.precedingFpeInnerTest != nil && fpeInnerTest.precedingFpeInnerTest.behindDescendantAxis:
			ok = true
			if doesUdpeTestMatches {
				next = fpeInnerTest.precedingFpeInnerTest
			} else {
				next = fpeInnerTest
			}
		case fpeInnerTest.behindDescendantAxis:
			ok = true
			next = fpeInnerTest
			if doesUdpeTestMatches {
				newTest = fpeInnerTest.precedingFpeInnerTest
				hasNewTest = true
			}
		case doesUdpeTestMatches:
			ok = true
			next = fpeInnerTest.precedingFpeInnerTest
		default:
			ok = false
		}
	}
	if ok {
		predicate = fpeInnerTest.udpeTest.predicate()
	}
	return
}

func (fpeInnerTest *fpeInnerTestImpl) entry() bool {
	return fpeInnerTest.isEntry
}

func (fpeInnerTest *fpeInnerTestImpl) String() (result string) {
	result = fmt.Sprintf("%v", fpeInnerTest.udpeTest)
	if fpeInnerTest.behindDescendantAxis {
		result += "//"
	} else if !fpeInnerTest.isEntry {
		result += "/"
	}
	return
}

//concrete fpe path pattern implementation
type fpePathPatternImpl struct {
	currentTest *fpeInnerTestImpl
}

func (fpePP *fpePathPatternImpl) isEmpty() bool {
	return fpePP.currentTest == nil
}

func (fpePP *fpePathPatternImpl) matchWithReductionOf(n interface{}, doUpdate bool) (predicate predicate, newPathPattern pathPattern, ok bool) {
	if fpePP.isEmpty() {
		panic(`fpe path pattern error: trying a match for an empty path pattern`)
	}

	predicate, next, newTest, hasNewTest, ok := fpePP.currentTest.matchWithReductionOf(n)
	if !ok {
		return
	}

	if !doUpdate {
		return
	}

	fpePP.currentTest = next
	if hasNewTest {
		newPathPattern = &fpePathPatternImpl{newTest}
	}
	return
}

func (fpePP *fpePathPatternImpl) String() (result string) {
	result = fpeStringifyUtil(fpePP.currentTest, fpePathPatternMode)

	if !strings.HasSuffix(result, "//") {
		result = strings.TrimSuffix(result, "/")
	}
	return
}

//concrete fpe implementation
type fpeImpl struct {
	entryTest *fpeInnerTestImpl
}

func (fpe *fpeImpl) entryPoint() pathPattern {
	return &fpePathPatternImpl{fpe.entryTest}
}

func (fpe *fpeImpl) String() string {
	return fpeStringifyUtil(fpe.entryTest, fpeMode)
}

//Utils
type fpeStringificationMode int

const (
	fpeMode fpeStringificationMode = iota
	fpePathPatternMode
)

func fpeStringifyUtil(currentTest *fpeInnerTestImpl, mode fpeStringificationMode) (result string) {
	if currentTest != nil {
		result = fpeStringifyUtil(currentTest.precedingFpeInnerTest, mode) + fmt.Sprintf("%v", currentTest)
	} else {
		if mode == fpePathPatternMode {
			result = "ε/"
		}
	}
	return
}
