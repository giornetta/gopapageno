package xpath

import (
	"testing"
)

func TestExecutionThreadList(t *testing.T) {

	t.Run(`newExecutionThreadList`, func(t *testing.T) {
		executionThreadList := newExecutionThreadList()
		want := 0
		if got := executionThreadList.len(); got != want {
			t.Errorf(`executionThreadList.len()=%v | want %v`, got, want)
		}
	})

	t.Run(`addExecutionThread(ctx, sol, pp)`, func(t *testing.T) {
		executionThreadList := newExecutionThreadList()
		executionThreadList.addExecutionThread(nil, nil, nil)
		want := 1
		if got := executionThreadList.len(); got != want {
			t.Errorf(`executionThreadList.len()=%v | want %v`, got, want)
		}
	})

	t.Run(`removeExecutionThread(executionThread, removeChildren)`, func(t *testing.T) {

		t.Run(`removeExecutionThread(executionThread, false)`, func(t *testing.T) {
			executionThreadList := newExecutionThreadList()
			executionThreadList.addExecutionThread(nil, nil, nil)
			removedExecutionThread := executionThreadList.addExecutionThread(nil, nil, nil)

			executionThreadList.removeExecutionThread(removedExecutionThread, false)
			want := 1
			if got := executionThreadList.len(); got != want {
				t.Errorf(`executionThreadList.len()=%v | want %v`, got, want)
			}
		})

		t.Run(`removeExecutionThread(executionThread, true)`, func(t *testing.T) {
			executionThreadList := newExecutionThreadList()
			executionThreadList.addExecutionThread(nil, nil, nil)
			superParentExecutionThread := executionThreadList.addExecutionThread(nil, nil, nil)
			superParentExecutionThread.addChild(executionThreadList.addExecutionThread(nil, nil, nil))
			superParentExecutionThread.addChild(executionThreadList.addExecutionThread(nil, nil, nil))
			parentExecutionThread := executionThreadList.addExecutionThread(nil, nil, nil)
			parentExecutionThread.addChild(executionThreadList.addExecutionThread(nil, nil, nil))
			superParentExecutionThread.addChild(parentExecutionThread)

			executionThreadList.removeExecutionThread(superParentExecutionThread, true)
			want := 1
			if got := executionThreadList.len(); got != want {
				t.Errorf(`executionThreadList.len()=%v | want %v`, got, want)
			}
		})
	})

	t.Run(`merge(incoming)`, func(t *testing.T) {
		t.Run(`merge(nil)`, func(t *testing.T) {
			executionThreadList := newExecutionThreadList()
			if result, ok := executionThreadList.merge(nil); result != executionThreadList || ok {
				t.Error(`merge(nil) returns incorrectly`)
			}
		})

		t.Run(`merge(incomingETList)`, func(t *testing.T) {
			executionThreadList := newExecutionThreadList()
			executionThreadList.addExecutionThread(nil, nil, nil)
			executionThreadList.addExecutionThread(nil, nil, nil)

			incomingExecutionThreadList := newExecutionThreadList()
			incomingExecutionThreadList.addExecutionThread(nil, nil, nil)

			if result, ok := executionThreadList.merge(incomingExecutionThreadList); result != executionThreadList || !ok {
				t.Error(`merge(incomingETList) returns incorrectly`)
			} else {
				const want = 3
				if got := result.len(); got != want {
					t.Errorf(`result.len()=%d | want %d`, got, want)
				}

				//TODO: check that all the execution threads have the right
				//executionThread.el field that connects them to the belonging list
			}
		})
	})

	t.Run(`hasExecutionThreadRunningFor(ctx)`, func(t *testing.T) {
		t.Run(`hasExecutionThreadRunningFor(ctx)=false`, func(t *testing.T) {
			executionThreadList := newExecutionThreadList()
			executionThreadList.addExecutionThread(nil, nil, nil)
			executionThreadList.addExecutionThread(nil, nil, nil)

			if found := executionThreadList.hasExecutionThreadRunningFor(NewNonTerminal()); found {
				t.Error(`hasExecutionThreadRunningFor(ctx)=true | want false`)
			}
		})
		t.Run(`hasExecutionThreadRunningFor(ctx)=true`, func(t *testing.T) {
			executionThreadList := newExecutionThreadList()
			findableContext := NewNonTerminal()
			executionThreadList.addExecutionThread(nil, nil, nil)
			executionThreadList.addExecutionThread(nil, nil, nil)
			executionThreadList.addExecutionThread(findableContext, nil, nil)
			executionThreadList.addExecutionThread(nil, nil, nil)

			if found := executionThreadList.hasExecutionThreadRunningFor(findableContext); !found {
				t.Error(`hasExecutionThreadRunningFor(ctx)=false | want true`)
			}
		})
	})
}

func TestExecutionThread(t *testing.T) {
	t.Run(`executionThread.isCompleted()`, func(t *testing.T) {
		t.Run(`executionThread.isCompleted()=true`, func(t *testing.T) {

			var et executionThread = &executionThreadImpl{
				pp: _emptyPathPatternBuilder(),
			}

			want := true
			if got := et.isCompleted(); got != want {
				t.Errorf(`executionThread.isCompleted()=%v | want %v`, got, want)
			}
		})

		t.Run(`executionThread.isCompleted()=false`, func(t *testing.T) {
			fpeBuilder := newFpeBuilder()
			fpeBuilder.addAxis(child)
			fpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
			fpe := fpeBuilder.end()
			nonEmptyPathPattern := fpe.entryPoint()

			var et executionThread = &executionThreadImpl{
				pp:     nonEmptyPathPattern,
				spList: newSpeculationList(),
			}

			want := false
			if got := et.isCompleted(); got != want {
				t.Errorf(`executionThread.isCompleted()=%v | want %v`, got, want)
			}
		})
	})

	t.Run(`executionThread.isSpeculative()`, func(t *testing.T) {
		t.Run(`executionThread.isSpeculative()=false`, func(t *testing.T) {
			var et executionThread = &executionThreadImpl{
				pp:     _emptyPathPatternBuilder(),
				spList: newSpeculationList(),
			}

			const want = false
			if got := et.isSpeculative(); got != want {
				t.Errorf(`executionThread.isSpeculative()=%v | want %v`, got, want)
			}
		})

		t.Run(`executionThread.isSpeculative()=true`, func(t *testing.T) {
			speculationList := newSpeculationList()
			speculationList.addSpeculation(new(predicateImpl), NewNonTerminal())

			var et executionThread = &executionThreadImpl{
				pp:     _emptyPathPatternBuilder(),
				spList: speculationList,
			}

			const want = true
			if got := et.isSpeculative(); got != want {
				t.Errorf(`executionThread.isSpeculative()=%v | want %v`, got, want)
			}
		})
	})
}

func TestExecutionThreadListIterator(t *testing.T) {
	t.Run(`iterator of empty execution thread list`, func(t *testing.T) {
		executionThreadList := newExecutionThreadList()
		executionThreadListIterator := executionThreadList.newIterator()

		t.Run(`hasNext()=false`, func(t *testing.T) {
			want := false
			if got := executionThreadListIterator.hasNext(); got != want {
				t.Errorf(`hasNext()=%v | want %v`, got, want)
			}
		})

		t.Run(`next() panics`, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error(`next() does NOT panic when it should`)
				}
			}()
			executionThreadListIterator.next()
		})
	})

	t.Run(`itearator of a NON empty execution thread list`, func(t *testing.T) {
		t.Run(`it allows right number of iterations`, func(t *testing.T) {
			executionThreadList := newExecutionThreadList()
			executionThreadList.addExecutionThread(nil, nil, nil)
			executionThreadList.addExecutionThread(nil, nil, nil)
			executionThreadList.addExecutionThread(nil, nil, nil)
			executionThreadList.addExecutionThread(nil, nil, nil)
			const expectedIterationCount = 4

			executionThreadListIterator := executionThreadList.newIterator()
			var actualIterationCount int
			for executionThreadListIterator.hasNext() {
				actualIterationCount++

				et, _ := executionThreadListIterator.next()
				if et == nil {
					t.Errorf(`next() returns a <nil> execution thread`)
				}
			}

			if actualIterationCount != expectedIterationCount {
				t.Errorf(`iterator iterated %d times | want %d`, actualIterationCount, expectedIterationCount)
			}
		})
	})
}

// Utils
func _emptyPathPatternBuilder() pathPattern {
	fpeBuilder := newFpeBuilder()
	fpeBuilder.addAxis(child)
	fpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
	fpe := fpeBuilder.end()
	emptyPathPattern := fpe.entryPoint()
	emptyPathPattern.matchWithReductionOf(newElement("a", nil, nil), true)

	return emptyPathPattern
}
