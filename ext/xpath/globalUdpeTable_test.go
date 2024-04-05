package xpath

import (
	"testing"
)

func TestGlobalUdpeTable(t *testing.T) {
	t.Run(`iterate(callback)`, func(t *testing.T) {
		udpeGlobalTable := new(globalUdpeTableImpl)

		udpeGlobalTable.addFpe(nil)
		udpeGlobalTable.addFpe(nil)
		udpeGlobalTable.addFpe(nil)

		const expectedCallCount = 3

		var actualCallCount int
		var spyCallback = func(id int, gur globalUdpeRecord) {
			actualCallCount++
		}

		udpeGlobalTable.iterate(spyCallback)

		if actualCallCount != expectedCallCount {
			t.Errorf(`callback called %d times | want %d times`, actualCallCount, expectedCallCount)
		}
	})
}
