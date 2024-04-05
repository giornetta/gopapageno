package xpath

type globalTableIterableCallback func(id int, record globalUdpeRecord)

type globalUdpeTable interface {
	newExecutionTable() executionTable
	addFpe(fpe fpe) (id int, record globalUdpeRecord)
	addRpe(rpe rpe) (id int, record globalUdpeRecord)
	iterate(callback globalTableIterableCallback)
	mainQueryRecord() globalUdpeRecord
}

type globalUdpeTableImpl struct {
	list []globalUdpeRecord
}

func (globalUdpeTable *globalUdpeTableImpl) newExecutionTable() executionTable {
	et := new(executionTableImpl)
	globalUdpeTableSize := globalUdpeTable.size()
	executionRecordsGroup := make([]executionRecord, globalUdpeTableSize)
	for id := range executionRecordsGroup {
		globalUdpeRecord := globalUdpeTable.recordByID(id)
		executionRecordsGroup[id] = &executionRecordImpl{
			expType:      globalUdpeRecord.udpeType(),
			t:            et,
			ctxSols:      newContextSolutionsMap(),
			etList:       newExecutionThreadList(),
			gNudpeRecord: globalUdpeRecord.nudpeRecord(),
		}
	}
	et.list = executionRecordsGroup

	return et
}

func (globalUdpeTable *globalUdpeTableImpl) size() int {
	return len(globalUdpeTable.list)
}

func (globalUdpeTable *globalUdpeTableImpl) recordByID(id int) globalUdpeRecord {
	return globalUdpeTable.list[id]
}

func (globalUdpeTable *globalUdpeTableImpl) mainQueryRecord() globalUdpeRecord {
	return globalUdpeTable.list[globalUdpeTable.size()-1]
}

func (globalUdpeTable *globalUdpeTableImpl) iterate(callback globalTableIterableCallback) {
	for id, gr := range globalUdpeTable.list {
		callback(id, gr)
	}
}

func (globalUdpeTable *globalUdpeTableImpl) addFpe(fpe fpe) (id int, record globalUdpeRecord) {
	return globalUdpeTable.addUdpe(fpe, FPE)
}

func (globalUdpeTable *globalUdpeTableImpl) addRpe(rpe rpe) (id int, record globalUdpeRecord) {
	return globalUdpeTable.addUdpe(rpe, RPE)
}

// addUdpe creates a new record inside the global udpe table
func (globalUdpeTable *globalUdpeTableImpl) addUdpe(udpe udpe, udpeType udpeType) (id int, record globalUdpeRecord) {
	r := &globalUdpeRecordImpl{
		exp:     udpe,
		expType: udpeType,
	}
	globalUdpeTable.list = append(globalUdpeTable.list, r)
	return len(globalUdpeTable.list) - 1, r
}

type globalUdpeRecord interface {
	udpe() udpe
	udpeType() udpeType
	nudpeRecord() globalNudpeRecord
	setNudpeRecord(nudpeRecord globalNudpeRecord)
}

type globalUdpeRecordImpl struct {
	exp          udpe
	expType      udpeType
	gNudpeRecord globalNudpeRecord
}

func (globalUdpeRecord *globalUdpeRecordImpl) udpe() udpe {
	return globalUdpeRecord.exp

}

// udpeType returns the type of the underlying UDPE
func (globalUdpeRecord *globalUdpeRecordImpl) udpeType() udpeType {
	return globalUdpeRecord.expType
}

// nudpeRecord returns the globalNudpeRecord of the NUDPE to which the UDPE belongs
func (globalUdpeRecord *globalUdpeRecordImpl) nudpeRecord() globalNudpeRecord {
	return globalUdpeRecord.gNudpeRecord
}

func (globalUdpeRecord *globalUdpeRecordImpl) setNudpeRecord(nudpeRecord globalNudpeRecord) {
	globalUdpeRecord.gNudpeRecord = nudpeRecord
}
