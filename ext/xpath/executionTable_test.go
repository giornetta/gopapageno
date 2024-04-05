package xpath

import "testing"

/*
	func TestExecutionTable(t *testing.T) {
		t.Run(`iterate(callback)`, func(t *testing.T) {
			const expectedCallcount = 5
			ExecutionTable := newExecutionTable(expectedCallcount)

			var actualCallcount int
			var spyCallback = func(id int, er executionRecord) (doBreak bool) {
				actualCallcount++
				if er == nil {
					t.Errorf(`callback called with a <nil> execution record`)
				}
				return
			}
			ExecutionTable.iterate(spyCallback)

			if actualCallcount != expectedCallcount {
				t.Errorf(`callback called %d times | want %d times`, actualCallcount, expectedCallcount)
			}
		})
	}
*/
func TestExecutionRecord(t *testing.T) {
	//TODO:
}
