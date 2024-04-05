package xpath

import (
	"fmt"
)

type executionTableIterableCallback func(id int, er executionRecord) (doBreak bool)

type executionTable interface {
	merge(incoming executionTable) (result executionTable, ok bool)
	iterate(callback executionTableIterableCallback)
	recordByID(id int) (executionRecord, error)
	mainQueryRecord() executionRecord
	evaluateID(udpeID int, context NonTerminal, evaluationsCount int) customBool
	size() int
}

type executionTableImpl struct {
	list []executionRecord
}

func (et *executionTableImpl) iterate(callback executionTableIterableCallback) {
	for id, er := range et.list {
		callback(id, er)
	}
}

func (et *executionTableImpl) recordByID(id int) (execRecord executionRecord, err error) {
	defer func() {
		if r := recover(); r != nil {
			execRecord = nil
			err = fmt.Errorf("execution table lookup: can NOT get execution record with id %d", id)
			return
		}
	}()
	execRecord = et.list[id]
	return
}

func (et *executionTableImpl) mainQueryRecord() executionRecord {
	return et.list[et.size()-1]
}

// merge joins the incoming execution table to the receiving execution table and returns
// the receiving execution table
func (et *executionTableImpl) merge(incoming executionTable) (result executionTable, ok bool) {
	result = et
	ok = true

	et.iterate(func(id int, er executionRecord) (doBreak bool) {
		incomingRecord, err := incoming.recordByID(id)
		if err != nil {
			ok = false
			doBreak = true
			return
		}

		if _, isMerged := er.merge(incomingRecord); !isMerged {
			ok = false
			doBreak = true
			return
		}
		return
	})
	return
}

func (et *executionTableImpl) size() int {
	return len(et.list)
}

// evaluateID returns the boolean value of an udpe with a certain id w.r.t a specific context.
func (et *executionTableImpl) evaluateID(udpeID int, context NonTerminal, evaluationsCount int) customBool {
	record, err := et.recordByID(udpeID)
	if err != nil {
		panic(fmt.Sprintf(`udpe boolean evaluation error for id: %v`, err))
	}

	//Do not consider the effective boolean value of all those UDPEs which are
	//the initial UDPE of some NUDPE.
	//The boolean value of a NUDPE is computed at the end of the entire parsing of the document.
	if record.belongsToNudpe() {
		return Undefined
	}

	if record.hasSolutionsFor(context) {
		return True
	}

	//FPEs are always bounded w.r.t. a specific context and their boolean value
	//is always available as soon as the context has been synthesized.
	if record.udpeType() == FPE {
		return False
	}

	//RPEs may be unbounded w.r.t. a certain context and their boolean values
	//may be undefined at the time when the context is synthesized.
	if evaluationsCount == 0 || record.hasExecutionThreadRunningFor(context) {
		return Undefined
	}

	return False
}

type executionRecord interface {
	addExecutionThread(ctx, sol NonTerminal, pp pathPattern) executionThread
	removeExecutionThread(et executionThread, removeChildren bool) (ok bool)
	hasExecutionThreadRunningFor(ctx NonTerminal) bool
	hasSolutionsFor(ctx NonTerminal) bool
	contextSolutions() contextSolutionsMap
	merge(incoming executionRecord) (result executionRecord, ok bool)
	updateAllExecutionThreads(reduced NonTerminal)
	stopUnfoundedSpeculativeExecutionThreads(evaluator evaluator)
	saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads(NonTerminal)
	produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads()
	udpeType() udpeType
	nudpeRecord() globalNudpeRecord
	belongsToNudpe() bool
}

type executionRecordImpl struct {
	expType      udpeType
	t            executionTable
	ctxSols      contextSolutionsMap
	etList       executionThreadList
	gNudpeRecord globalNudpeRecord
}

func (er *executionRecordImpl) String() string {
	return fmt.Sprintf("{ %v | %v | %v }", er.expType, er.ctxSols, er.etList)
}

func (er *executionRecordImpl) addExecutionThread(ctx, sol NonTerminal, pp pathPattern) (et executionThread) {
	et = er.etList.addExecutionThread(ctx, sol, pp)
	logger.Printf("adding execution thread: %v", et)
	return
}

func (er *executionRecordImpl) removeExecutionThread(et executionThread, removeChildren bool) (ok bool) {
	return er.etList.removeExecutionThread(et, removeChildren)
}

func (er *executionRecordImpl) hasExecutionThreadRunningFor(ctx NonTerminal) bool {
	return er.etList.hasExecutionThreadRunningFor(ctx)
}

func (er *executionRecordImpl) hasSolutionsFor(ctx NonTerminal) bool {
	return er.ctxSols.hasSolutionsFor(ctx)
}

func (er *executionRecordImpl) contextSolutions() contextSolutionsMap {
	return er.ctxSols
}

func (er *executionRecordImpl) udpeType() udpeType {
	return er.expType
}

func (er *executionRecordImpl) nudpeRecord() globalNudpeRecord {
	return er.gNudpeRecord
}

func (er *executionRecordImpl) belongsToNudpe() bool {
	return er.gNudpeRecord != nil
}

// updateExecutionThreads takes the node being reduced and asks all the running execution threads
// to update accordingly
func (er *executionRecordImpl) updateAllExecutionThreads(reduced NonTerminal) {
	er.etList.iterate(func(et executionThread) (doBreak bool) {
		etPathPattern := et.pathPattern()
		//The path pattern of the execution thread may be empty if the thread is speculative
		//and it's not completed because of some unchecked speculation
		if etPathPattern.isEmpty() {
			return
		}

		etReprBeforeUpdate := fmt.Sprintf("%v", et)
		predicate, newPathPattern, ok := etPathPattern.matchWithReductionOf(reduced.Node(), true)
		if !ok {
			logger.Printf("removing execution thread beacuse path pattern does NOT match: %s", etReprBeforeUpdate)
			er.etList.removeExecutionThread(et, false)
			return
		}
		etReprAfterUpdate := fmt.Sprintf("%v", et)
		logger.Printf("updated execution thread: %s -> %s", etReprBeforeUpdate, etReprAfterUpdate)

		var etReceivingSpeculation = et
		if newPathPattern != nil {
			etReceivingSpeculation = er.addExecutionThread(et.context(), et.solution(), newPathPattern)
			et.addChild(etReceivingSpeculation)
		}

		if predicate != nil {
			sp := etReceivingSpeculation.addSpeculation(predicate, reduced)
			logger.Printf("adding speculation: %v to execution thread %v", sp, et)
		}
		return
	})
}

// stopUnfoundedSpeculativeExecutionThreads iterates over all the running execution threads and, for each speculative
// execution thread, the speculation is evaluated. If the speculation ends up to be unfounded, the speculative execution thread,
// and all its Children recursively, are stopped
func (er *executionRecordImpl) stopUnfoundedSpeculativeExecutionThreads(evaluator evaluator) {
	er.etList.iterate(func(execThread executionThread) (doBreak bool) {
		if areSpeculationsFounded := execThread.checkAndUpdateSpeculations(evaluator); !areSpeculationsFounded {
			if isExecutionThreadRemoved := er.etList.removeExecutionThread(execThread, true); !isExecutionThreadRemoved {
				panic("stopping unfounded speculative execution thred: cannot remove execution thread")
			}
		}
		return
	})
}

// saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads saves the input NonTerminal as either context or solution
// for all the execution threads that are completed. By the time at which the execution thread is completed, it might not be able
// to produce context-solution items because of running speculations. Even if the execution thread can not produce context-solution
// items, it has to save the non terminal whose reducetion caused the execution thread to complete
func (er *executionRecordImpl) saveReducedNTAsContextOrSolutionlIntoCompletedExecutionThreads(contextOrSolution NonTerminal) {
	er.etList.iterate(func(execThread executionThread) (doBreak bool) {
		if execThread.isCompleted() {
			execThread.setNTAsContextOrSolutionIfNotAlreadySet(contextOrSolution)
		}
		return
	})
}

// produceContextSolutions produce new context solutions from completed execution threads and removes
// completed execution threads
func (er *executionRecordImpl) produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads() {
	er.etList.iterate(func(et executionThread) (doBreak bool) {
		if et.isCompleted() && !et.isSpeculative() {
			logger.Printf("adding context-solution: (%v , %v)", et.context(), et.solution())
			er.ctxSols.addContextSolution(et.context(), et.solution())
			er.etList.removeExecutionThread(et, false)
		}
		return
	})
}

func (er *executionRecordImpl) merge(incoming executionRecord) (result executionRecord, ok bool) {
	result = er
	incomingImpl, ok := incoming.(*executionRecordImpl)

	if !ok {
		return
	}

	_, okMergeEtLists := er.etList.merge(incomingImpl.etList)
	_, okMergeCtxSols := er.ctxSols.merge(incomingImpl.ctxSols)

	ok = okMergeEtLists && okMergeCtxSols
	return
}
