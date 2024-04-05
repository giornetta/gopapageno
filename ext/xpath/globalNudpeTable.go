package xpath

type globalNudpeTable interface {
	recordByID(id int) globalNudpeRecord
	mainQueryRecord() globalNudpeRecord
	addNudpeRecord(length int) globalNudpeRecord
	size() int
	newIterator() globalNudpeTableIterator
}

type globalNudpeTableImpl struct {
	list []*globalNudpeRecordImpl
}

func (globalNudpeTable *globalNudpeTableImpl) recordByID(id int) globalNudpeRecord {
	return globalNudpeTable.list[id]
}

func (globalNudpeTable *globalNudpeTableImpl) mainQueryRecord() globalNudpeRecord {
	return globalNudpeTable.list[globalNudpeTable.size()-1]
}

func (globalNudpeTable *globalNudpeTableImpl) addNudpeRecord(length int) globalNudpeRecord {
	newNudpeRecord := &globalNudpeRecordImpl{
		len:     length,
		ctxSols: newContextSolutionsMap(),
	}
	globalNudpeTable.list = append(globalNudpeTable.list, newNudpeRecord)
	return newNudpeRecord
}

// size returns the number of NUDPE which are recorded inside the table
func (globalNudpeTable *globalNudpeTableImpl) size() int {
	return len(globalNudpeTable.list)
}

func (globalNudpeTable *globalNudpeTableImpl) newIterator() globalNudpeTableIterator {
	return &globalNudpeTableIteratorImpl{
		table: globalNudpeTable,
	}
}

type globalNudpeRecord interface {
	length() int
	hasSolutionsFor(ctx NonTerminal) bool
	contextSolutions() contextSolutionsMap
	setContextSolutions(contextSolutionsMap)
}

type globalNudpeRecordImpl struct {
	ctxSols contextSolutionsMap
	len     int
}

func (globalNudpeRecord *globalNudpeRecordImpl) contextSolutions() contextSolutionsMap {
	return globalNudpeRecord.ctxSols
}

func (globalNudpeRecord *globalNudpeRecordImpl) setContextSolutions(ctxSols contextSolutionsMap) {
	globalNudpeRecord.ctxSols = ctxSols
}

// length returns the number of UDPE by which the NUDPE is composed
func (globalNudpeRecord *globalNudpeRecordImpl) length() int {
	return globalNudpeRecord.len
}

func (globalNudpeRecord *globalNudpeRecordImpl) hasSolutionsFor(ctx NonTerminal) bool {
	return globalNudpeRecord.ctxSols.hasSolutionsFor(ctx)
}

type globalNudpeTableIterator interface {
	hasNext() bool
	next() globalNudpeRecord
}

type globalNudpeTableIteratorImpl struct {
	table        *globalNudpeTableImpl
	nextRecordID int
}

func (iterator *globalNudpeTableIteratorImpl) hasNext() bool {
	return iterator.nextRecordID < len(iterator.table.list)-1
}

func (iterator *globalNudpeTableIteratorImpl) next() globalNudpeRecord {
	defer func() { iterator.nextRecordID++ }()
	return iterator.table.list[iterator.nextRecordID]
}
