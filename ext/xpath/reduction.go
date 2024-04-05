package xpath

import (
	"fmt"
)

type Reduction struct {
	reducedNT, generativeNT, wrappedNT NonTerminal
	updatingExecutionTable             executionTable
	globalUdpeRecordBeingConsidered    globalUdpeRecord
}

func (r *Reduction) Setup(reducedNT, generativeNT, wrappedNT NonTerminal) {
	var updatingExecutionTable executionTable

	if wrappedNT != nil {
		updatingExecutionTable = wrappedNT.ExecutionTable()
	} else {
		updatingExecutionTable = udpeGlobalTable.newExecutionTable()
	}

	r.reducedNT = reducedNT
	r.generativeNT = generativeNT
	r.wrappedNT = wrappedNT
	r.updatingExecutionTable = updatingExecutionTable
}

func (r *Reduction) Reset() {
	r.reducedNT = nil
	r.generativeNT = nil
	r.wrappedNT = nil
	r.updatingExecutionTable = nil
	r.globalUdpeRecordBeingConsidered = nil
}

func (r *Reduction) Handle() {
	defer r.avoidMemoryLeaksAtTheEndOfHandling()

	r.iterateOverAllGlobalUdpeRecordsAndExecuteMainPhases()
	r.prepareUpdatingExecutionTableToBePropagatedToReducedNT()
	r.propagateUpdatingExecutionTableToReducedNT()
}

func (r *Reduction) avoidMemoryLeaksAtTheEndOfHandling() {
	r.reducedNT = nil
	r.generativeNT = nil
	r.wrappedNT = nil
	r.updatingExecutionTable = nil
	r.globalUdpeRecordBeingConsidered = nil
}

func (r *Reduction) iterateOverAllGlobalUdpeRecordsAndExecuteMainPhases() {
	udpeGlobalTable.iterate(func(id int, globalRecord globalUdpeRecord) {
		r.globalUdpeRecordBeingConsidered = globalRecord
		updatingExecutionRecord, err := r.updatingExecutionTable.recordByID(id)
		if err != nil {
			panic(fmt.Sprintf("cannot retrieve execution record for udpe with id: %d", id))
		}

		r.addNewExecutionThreadsToExecutionRecord(updatingExecutionRecord)
		updatingExecutionRecord.updateAllExecutionThreads(r.reducedNT)
		updatingExecutionRecord.stopUnfoundedSpeculativeExecutionThreads(r.updatingExecutionTable.evaluateID)
		updatingExecutionRecord.saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads(r.reducedNT)
		updatingExecutionRecord.produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads()
	})
}

func (r *Reduction) prepareUpdatingExecutionTableToBePropagatedToReducedNT() {
	if r.generativeNT != nil {
		r.mergeUpdatingExecutionTableWithUnchangedExecutionTable()
	}
}

func (r *Reduction) propagateUpdatingExecutionTableToReducedNT() {
	r.reducedNT.SetExecutionTable(r.updatingExecutionTable)
}

func (r *Reduction) mergeUpdatingExecutionTableWithUnchangedExecutionTable() {
	unchangedExecutionTable := r.generativeNT.ExecutionTable()
	_, ok := r.updatingExecutionTable.merge(unchangedExecutionTable)
	if !ok {
		panic(`Reduction Handle Node error: can NOT merge execution tables`)
	}
}

// phase1 checks if the udpe's entry point matches current Reduction, without updating the path pattern,
// and, if it maches, creates new execution threads accordingly
func (r *Reduction) addNewExecutionThreadsToExecutionRecord(executionRecord executionRecord) {
	udpe := r.globalUdpeRecordBeingConsidered.udpe()
	entryPoint := udpe.entryPoint()

	if _, _, ok := entryPoint.matchWithReductionOf(r.reducedNT.Node(), false); !ok {
		return
	}

	switch udpeType := r.globalUdpeRecordBeingConsidered.udpeType(); udpeType {
	case FPE:
		executionRecord.addExecutionThread(nil, r.reducedNT, entryPoint)
	case RPE:
		if r.wrappedNT == nil {
			return
		}
		executionRecord.addExecutionThread(r.wrappedNT, nil, entryPoint)
		childrenOfWrappedNT := r.wrappedNT.Children()
		for _, child := range childrenOfWrappedNT {
			executionRecord.addExecutionThread(child, nil, udpe.entryPoint())
		}

	default:
		panic(fmt.Sprintf(`adding new execution threads to execution record: unknown udpe type %q`, udpeType))
	}
}
