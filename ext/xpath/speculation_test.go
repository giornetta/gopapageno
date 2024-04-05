package xpath

import (
	"testing"
)

func TestSpeculationList(t *testing.T) {
	t.Run(`newSpeculationList`, func(t *testing.T) {
		speculationList := newSpeculationList()
		want := 0
		if got := speculationList.len(); got != want {
			t.Errorf(`specualationList.len()=%v | want %v`, got, want)
		}
	})

	t.Run(`speculationList.addSpeculation(predicate, context)`, func(t *testing.T) {
		speculationList := newSpeculationList()
		speculationList.addSpeculation(nil, nil)
		want := 1
		if got := speculationList.len(); got != want {
			t.Errorf(`speculationList.len()=%d | want %d`, got, want)
		}
	})

	t.Run(`speculationList.removeSpeculation(speculation`, func(t *testing.T) {
		speculationList := newSpeculationList()
		sp := speculationList.addSpeculation(nil, nil)
		speculationList.removeSpeculation(sp)
		want := 0
		if got := speculationList.len(); got != 0 {
			t.Errorf(`speculationList.len()=%d | want %d`, got, want)
		}

	})

	t.Run(`speculationList.len()`, func(t *testing.T) {
		speculationList := newSpeculationList()
		want := 0
		if got := speculationList.len(); got != want {
			t.Errorf(`speculationList.len()=%d | want %d`, got, want)
		}

		speculationList.addSpeculation(nil, nil)
		speculationList.addSpeculation(nil, nil)
		speculationList.addSpeculation(nil, nil)
		want = 3
		if got := speculationList.len(); got != want {
			t.Errorf(`speculationList.len()=%d | want %d`, got, want)
		}

	})
}

func TestSpeculationListIterator(t *testing.T) {
	t.Run(`iterator of empty speculation list`, func(t *testing.T) {
		speculationList := newSpeculationList()
		speculationListIterator := speculationList.newIterator()

		t.Run(`speculationListIterator.hasNext()=false`, func(t *testing.T) {
			want := false
			if got := speculationListIterator.hasNext(); got != want {
				t.Errorf(`speculationListIterator.hasNext()=%v | want %v`, got, want)
			}
		})

		t.Run(`speculationListIterator.next() panics`, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf(`speculationListIterator.next() does NOT panic when it should`)
				}
			}()
			speculationListIterator.next()
		})
	})

	t.Run(`iterator of NON empty speculation list`, func(t *testing.T) {
		t.Run(`it allows right number of iterations`, func(t *testing.T) {
			speculationList := newSpeculationList()
			speculationList.addSpeculation(nil, nil)
			speculationList.addSpeculation(nil, nil)
			speculationList.addSpeculation(nil, nil)
			speculationList.addSpeculation(nil, nil)
			const expectedIterationCount = 4

			speculationListIterator := speculationList.newIterator()
			var actualIterationCount int

			for speculationListIterator.hasNext() {
				actualIterationCount++
				sp, _ := speculationListIterator.next()

				if sp == nil {
					t.Errorf(`speculationListIterator.next() returns a <nil> speculation`)
				}
			}

			if actualIterationCount != expectedIterationCount {
				t.Errorf(`iterator iterated %d times | want %d times`, actualIterationCount, expectedIterationCount)
			}
		})

		t.Run(`it supports speculation removal during iteration`, func(t *testing.T) {
			speculationList := newSpeculationList()
			speculationList.addSpeculation(nil, nil)
			speculationList.addSpeculation(nil, nil)
			speculationList.addSpeculation(nil, nil)
			speculationList.addSpeculation(nil, nil)
			const expectedIterationCount = 4

			const expectedSpAfterIteration = 3
			speculationListIterator := speculationList.newIterator()
			var actualIterationCount int

			for speculationListIterator.hasNext() {
				actualIterationCount++

				sp, _ := speculationListIterator.next()
				if actualIterationCount == 2 {
					speculationList.removeSpeculation(sp)
				}
			}

			if actualIterationCount != expectedIterationCount {
				t.Errorf(`iterator iterated %d times | want %d times`, actualIterationCount, expectedIterationCount)
			}

			if actualSpAfterIteration := speculationList.len(); actualSpAfterIteration != expectedSpAfterIteration {
				t.Errorf(`speculationList.len()=%d | want %d`, actualSpAfterIteration, expectedSpAfterIteration)
			}
		})
	})
}
