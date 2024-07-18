package xpath

import (
	"context"
	"fmt"
	"github.com/giornetta/gopapageno"
	"log"
	"os"
)

const defaultExecutorNumberOfThreads = 1

// Singletons
var nudpeGlobalTable globalNudpeTable
var udpeGlobalTable globalUdpeTable
var logger Logger

func init() {
	logger = newNopLogger()
}

type mainQueryType int

const (
	UDPE mainQueryType = iota
	NUDPE
)

type executor interface {
	setXPathQueryToBeExecuted(xpathquery string)
	setNumberOfThreadsToBeUsedToParseDocument(numberOfThreads int)
	setDocumentToBeParsedFilePath(documentFilePath string)
	initSingletonDataStructures()
	freeSingletonDataStructures()
	parseXPathQueryAndPopulateSingletonsDataStructures() (err error)
	executeUDPEsWhileParsingDocumentFile() (err error)
	completeExecutionOfUDPEsAndNUDPEs() error
	retrieveResults() []Position
}

type executorImpl struct {
	numberOfThreads         int
	xpathQueryToBeExecuted  string
	mainQueryType           mainQueryType
	resultingExecutionTable executionTable

	source []byte
	runner *gopapageno.Runner
}

// ExecutorCommand represents a command that can be made by a
// client to execute a XPath query
type ExecutorCommand interface {
	Execute(xpathQuery string) ExecutorCommand
	Against(source []byte) ExecutorCommand
	WithNumberOfThreads(numberOfThreads int) ExecutorCommand
	Run(runner *gopapageno.Runner) (results []Position, err error)
	InVerboseMode() ExecutorCommand
}

type executorCommandImpl struct {
	xpathQuery      string
	source          []byte
	numberOfThreads int
	verbose         bool
}

// Execute specify the XPath query to be executed
func Execute(xpathQuery string) ExecutorCommand {
	return &executorCommandImpl{
		xpathQuery: xpathQuery,
	}
}

func (executorCommand *executorCommandImpl) Execute(xpathQuery string) ExecutorCommand {
	executorCommand.xpathQuery = xpathQuery
	return executorCommand
}

func (executorCommand *executorCommandImpl) Against(source []byte) ExecutorCommand {
	executorCommand.source = source
	return executorCommand
}

// WithNumberOfThreads specify the number of threads to be used
// to execute the XPath query
func WithNumberOfThreads(numberOfThreads int) ExecutorCommand {
	return &executorCommandImpl{
		numberOfThreads: numberOfThreads,
	}
}

func (executorCommand *executorCommandImpl) WithNumberOfThreads(numberOfThreads int) ExecutorCommand {
	executorCommand.numberOfThreads = numberOfThreads
	return executorCommand
}

func (executorCommand *executorCommandImpl) InVerboseMode() ExecutorCommand {
	executorCommand.verbose = true
	logger = log.New(os.Stderr, "", log.LstdFlags)
	return executorCommand
}

// Run takes care of executing the command and to return
func (executorCommand *executorCommandImpl) Run(runner *gopapageno.Runner) (results []Position, err error) {
	executor := &executorImpl{
		numberOfThreads: defaultExecutorNumberOfThreads,
		runner:          runner,
	}

	executor.initSingletonDataStructures()
	defer executor.freeSingletonDataStructures()

	executor.setXPathQueryToBeExecuted(executorCommand.xpathQuery)
	executor.setNumberOfThreadsToBeUsedToParseDocument(executorCommand.numberOfThreads)
	executor.source = executorCommand.source

	err = executor.parseXPathQueryAndPopulateSingletonsDataStructures()
	if err != nil {
		return
	}

	err = executor.executeUDPEsWhileParsing()
	if err != nil {
		return
	}

	err = executor.completeExecutionOfUDPEsAndNUDPEs()
	if err != nil {
		return
	}

	results = executor.retrieveResults()

	return
}

func (executor *executorImpl) initSingletonDataStructures() {
	udpeGlobalTable = new(globalUdpeTableImpl)
	nudpeGlobalTable = new(globalNudpeTableImpl)
}

func (executor *executorImpl) freeSingletonDataStructures() {
	udpeGlobalTable = nil
	nudpeGlobalTable = nil
}

func (executor *executorImpl) setXPathQueryToBeExecuted(xpathQuery string) {
	executor.xpathQueryToBeExecuted = xpathQuery
}

func (executor *executorImpl) setNumberOfThreadsToBeUsedToParseDocument(numberOfThreads int) {
	executor.numberOfThreads = numberOfThreads
}

func (executor *executorImpl) parseXPathQueryAndPopulateSingletonsDataStructures() (err error) {

	switch executor.xpathQueryToBeExecuted {
	case "A1":
		executor.A1()
	case "A2":
		executor.A2()
	case "A3":
		executor.A3()
	case "A4":
		executor.A4()
	case "A5":
		executor.A5()
	case "A6":
		executor.A6()
	case "A7":
		executor.A7()
	case "A8":
		executor.A8()
	case "B1":
		executor.B1()
	case "B2":
		executor.B2()
	default:
		return fmt.Errorf("unknown query: %s", executor.xpathQueryToBeExecuted)
	}
	return nil
}

func (executor *executorImpl) executeUDPEsWhileParsing() error {
	axiomToken, err := executor.runner.Run(context.Background(), executor.source)
	if err != nil {
		return fmt.Errorf("could not parse: %v", err)
	}
	executor.resultingExecutionTable = axiomToken.Value.(NonTerminal).ExecutionTable()

	return err
}

func (executor *executorImpl) completeExecutionOfUDPEsAndNUDPEs() (err error) {

	if numberOfNUDPEs := nudpeGlobalTable.size(); numberOfNUDPEs == 0 {
		return
	}

	var currentNudpeRecord globalNudpeRecord
	var currentNudpeContextSolutionsMaps []contextSolutionsMap

	executor.resultingExecutionTable.iterate(func(id int, er executionRecord) (doBreak bool) {
		if er.belongsToNudpe() {
			if er.nudpeRecord() != currentNudpeRecord {
				if currentNudpeRecord != nil {
					currentNudpeRecord.setContextSolutions(transitiveClosure(currentNudpeContextSolutionsMaps))
				}
				currentNudpeRecord = er.nudpeRecord()
				currentNudpeContextSolutionsMaps = []contextSolutionsMap{}
			}
		} else {
			if currentNudpeRecord != nil {
				currentNudpeRecord.setContextSolutions(transitiveClosure(currentNudpeContextSolutionsMaps))
				currentNudpeRecord = nil
				currentNudpeContextSolutionsMaps = nil
			}
		}

		er.stopUnfoundedSpeculativeExecutionThreads(executor.nudpeBooleanValueEvaluator)
		er.produceContextSolutionsOutOfCompletedNonSpeculativeExecutionThreads()

		if er.belongsToNudpe() {
			currentNudpeContextSolutionsMaps = append(currentNudpeContextSolutionsMaps, er.contextSolutions())
		}
		return
	})

	if currentNudpeRecord != nil {
		currentNudpeRecord.setContextSolutions(transitiveClosure(currentNudpeContextSolutionsMaps))
	}

	return
}

func (executor *executorImpl) nudpeBooleanValueEvaluator(udpeID int, context NonTerminal, evaluationsCount int) customBool {
	record, err := executor.resultingExecutionTable.recordByID(udpeID)

	if err != nil {
		panic(fmt.Sprintf(`cannot retrieve udpe record from resulting execution table: %v`, err))
	}

	if !record.belongsToNudpe() {
		panic("cannot evaluate nudpe boolean value if first udpe does NOT belong to a nudpe")
	}

	nudpeGlobalRecord := record.nudpeRecord()
	return toCustomBool(nudpeGlobalRecord.hasSolutionsFor(context))
}

func (executor *executorImpl) retrieveResults() (results []Position) {
	var potentiallyDuplicatedResults []Position
	switch executor.mainQueryType {
	case UDPE:
		mainQueryUDPERecord := executor.resultingExecutionTable.mainQueryRecord()
		potentiallyDuplicatedResults = mainQueryUDPERecord.contextSolutions().convertToGroupOfSolutionsPositions()
	case NUDPE:
		mainQueryNUDPERecord := nudpeGlobalTable.mainQueryRecord()
		potentiallyDuplicatedResults = mainQueryNUDPERecord.contextSolutions().convertToGroupOfSolutionsPositions()
	default:
		panic("retrieving results error: unknown main query type")
	}

	var uniqueResults = make(map[string]bool, len(potentiallyDuplicatedResults))

	for _, result := range potentiallyDuplicatedResults {
		resultIdentifier := fmt.Sprintf("(%v,%v)", result.Start(), result.End())
		_, isAlreadyConsidered := uniqueResults[resultIdentifier]
		if !isAlreadyConsidered {
			results = append(results, result)
			uniqueResults[resultIdentifier] = true
		}
	}

	return
}

func (executor *executorImpl) A1() {
	executor.mainQueryType = UDPE
	/* /site/closed_auctions/closed_auction/annotation/description/Text/keyword */

	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("site", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("closed_auctions", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("closed_auction", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("annotation", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("description", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("Text", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("keyword", nil, nil))

	udpeGlobalTable.addFpe(fpeBuilder1.end())
}

func (executor *executorImpl) A2() {
	executor.mainQueryType = UDPE
	/* //closed_auction//keyword */

	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(descendantOrSelf)
	fpeBuilder1.addUdpeTest(newElementTest("closed_auction", nil, nil))
	fpeBuilder1.addAxis(descendantOrSelf)
	fpeBuilder1.addUdpeTest(newElementTest("keyword", nil, nil))

	udpeGlobalTable.addFpe(fpeBuilder1.end())
}

func (executor *executorImpl) A3() {
	executor.mainQueryType = UDPE
	/* /site/closed_auctions/closed_auction//keyword */

	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("site", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("closed_auctions", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("closed_auction", nil, nil))
	fpeBuilder1.addAxis(descendantOrSelf)
	fpeBuilder1.addUdpeTest(newElementTest("keyword", nil, nil))

	udpeGlobalTable.addFpe(fpeBuilder1.end())
}

func (executor *executorImpl) A4() {
	executor.mainQueryType = UDPE

	/* p = /closed_auction/annotation/description/Text/keyword */

	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("closed_auction", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("annotation", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("description", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("Text", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("keyword", nil, nil))

	fpeID, _ := udpeGlobalTable.addFpe(fpeBuilder1.end())

	p := &predicateImpl{
		expressionVector: []operator{atom()},
		atomsLookup: map[atomID]int{
			atomID(fpeID): 0,
		},
	}

	/* /site/closed_auctions/closed_auction[p]/date */

	fpeBuilder2 := newFpeBuilder()
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("site", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("closed_auctions", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("closed_auction", nil, p))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("date", nil, nil))

	udpeGlobalTable.addFpe(fpeBuilder2.end())
}

func (executor *executorImpl) A5() {
	executor.mainQueryType = UDPE

	/* p = closed_auction//keyword */
	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("closed_auction", nil, nil))
	fpeBuilder1.addAxis(descendantOrSelf)
	fpeBuilder1.addUdpeTest(newElementTest("keyword", nil, nil))

	fpeID, _ := udpeGlobalTable.addFpe(fpeBuilder1.end())

	p := &predicateImpl{
		expressionVector: []operator{atom()},
		atomsLookup: map[atomID]int{
			atomID(fpeID): 0,
		},
	}

	/* /site/closed_auctions/closed_auction[p]/date */
	fpeBuilder2 := newFpeBuilder()
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("site", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("closed_auctions", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("closed_auction", nil, p))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("date", nil, nil))

	udpeGlobalTable.addFpe(fpeBuilder2.end())
}

func (executor *executorImpl) A6() {
	executor.mainQueryType = UDPE

	/* /person/profile/gender */
	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("profile", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("gender", nil, nil))

	/* /person/profile/age */
	fpeBuilder2 := newFpeBuilder()
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("profile", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("age", nil, nil))

	fpe1ID, _ := udpeGlobalTable.addFpe(fpeBuilder1.end())
	fpe2ID, _ := udpeGlobalTable.addFpe(fpeBuilder2.end())

	p := &predicateImpl{
		expressionVector: []operator{and(), atom(), atom()},
		atomsLookup: map[atomID]int{
			atomID(fpe1ID): 1,
			atomID(fpe2ID): 2,
		},
	}

	/* /site/people/person[p]/name */
	fpebuilder3 := newFpeBuilder()
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("site", nil, nil))
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("people", nil, nil))
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("person", nil, p))
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("name", nil, nil))

	udpeGlobalTable.addFpe(fpebuilder3.end())
}

func (executor *executorImpl) A7() {
	executor.mainQueryType = UDPE

	/* /person/phone */
	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("phone", nil, nil))

	/* /person/homepage */
	fpeBuilder2 := newFpeBuilder()
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("homepage", nil, nil))

	fpe1ID, _ := udpeGlobalTable.addFpe(fpeBuilder1.end())
	fpe2ID, _ := udpeGlobalTable.addFpe(fpeBuilder2.end())

	p := &predicateImpl{
		expressionVector: []operator{or(), atom(), atom()},
		atomsLookup: map[atomID]int{
			atomID(fpe1ID): 1,
			atomID(fpe2ID): 2,
		},
	}

	/* /site/people/person[p]/name */
	fpebuilder3 := newFpeBuilder()
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("site", nil, nil))
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("people", nil, nil))
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("person", nil, p))
	fpebuilder3.addAxis(child)
	fpebuilder3.addUdpeTest(newElementTest("name", nil, nil))

	udpeGlobalTable.addFpe(fpebuilder3.end())

}

func (executor *executorImpl) A8() {
	executor.mainQueryType = UDPE

	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder1.addAxis(child)
	fpeBuilder1.addUdpeTest(newElementTest("address", nil, nil))

	fpeBuilder2 := newFpeBuilder()
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("phone", nil, nil))

	fpeBuilder3 := newFpeBuilder()
	fpeBuilder3.addAxis(child)
	fpeBuilder3.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder3.addAxis(child)
	fpeBuilder3.addUdpeTest(newElementTest("homepage", nil, nil))

	fpeBuilder4 := newFpeBuilder()
	fpeBuilder4.addAxis(child)
	fpeBuilder4.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder4.addAxis(child)
	fpeBuilder4.addUdpeTest(newElementTest("creditcard", nil, nil))

	fpeBuilder5 := newFpeBuilder()
	fpeBuilder5.addAxis(child)
	fpeBuilder5.addUdpeTest(newElementTest("person", nil, nil))
	fpeBuilder5.addAxis(child)
	fpeBuilder5.addUdpeTest(newElementTest("profile", nil, nil))

	fpe1ID, _ := udpeGlobalTable.addFpe(fpeBuilder1.end())
	fpe2ID, _ := udpeGlobalTable.addFpe(fpeBuilder2.end())
	fpe3ID, _ := udpeGlobalTable.addFpe(fpeBuilder3.end())
	fpe4ID, _ := udpeGlobalTable.addFpe(fpeBuilder4.end())
	fpe5ID, _ := udpeGlobalTable.addFpe(fpeBuilder5.end())

	p := &predicateImpl{
		expressionVector: []operator{and(), atom(), and(), nil, nil, or(), or(), nil, nil, nil, nil, atom(), atom(), atom(), atom()},
		atomsLookup: map[atomID]int{
			atomID(fpe1ID): 1,
			atomID(fpe2ID): 11,
			atomID(fpe3ID): 12,
			atomID(fpe4ID): 13,
			atomID(fpe5ID): 14,
		},
	}

	/* /site/people/person[p]/name */
	fpebuilder6 := newFpeBuilder()
	fpebuilder6.addAxis(child)
	fpebuilder6.addUdpeTest(newElementTest("site", nil, nil))
	fpebuilder6.addAxis(child)
	fpebuilder6.addUdpeTest(newElementTest("people", nil, nil))
	fpebuilder6.addAxis(child)
	fpebuilder6.addUdpeTest(newElementTest("person", nil, p))
	fpebuilder6.addAxis(child)
	fpebuilder6.addUdpeTest(newElementTest("name", nil, nil))

	udpeGlobalTable.addFpe(fpebuilder6.end())
}

func (executor *executorImpl) B1() {
	executor.mainQueryType = UDPE

	rpeBuilder1 := newRpeBuilder()
	rpeBuilder1.addAxis(parent)
	rpeBuilder1.addUdpeTest(newElementTest("namerica", nil, nil))
	rpeBuilder1.addAxis(parent)
	rpeBuilder1.addUdpeTest(newElementTest("samerica", nil, nil))

	rpe1ID, _ := udpeGlobalTable.addRpe(rpeBuilder1.end())

	p := &predicateImpl{
		expressionVector: []operator{atom()},
		atomsLookup: map[atomID]int{
			atomID(rpe1ID): 0,
		},
	}

	/* /site/regions/namerica/item[p]/name */
	fpebuilder2 := newFpeBuilder()
	fpebuilder2.addAxis(child)
	fpebuilder2.addUdpeTest(newElementTest("site", nil, nil))
	fpebuilder2.addAxis(child)
	fpebuilder2.addUdpeTest(newElementTest("regions", nil, nil))
	fpebuilder2.addAxis(child)
	fpebuilder2.addUdpeTest(newElementTest("*", nil, nil))
	fpebuilder2.addAxis(child)
	fpebuilder2.addUdpeTest(newElementTest("item", nil, p))
	fpebuilder2.addAxis(child)
	fpebuilder2.addUdpeTest(newElementTest("name", nil, nil))

	udpeGlobalTable.addFpe(fpebuilder2.end())
}

func (executor *executorImpl) B2() {
	executor.mainQueryType = NUDPE
	nudpeRec := nudpeGlobalTable.addNudpeRecord(3)

	fpeBuilder1 := newFpeBuilder()
	fpeBuilder1.addAxis(descendantOrSelf)
	fpeBuilder1.addUdpeTest(newElementTest("keyword", nil, nil))

	rpeBuilder1 := newRpeBuilder()
	rpeBuilder1.addAxis(ancestorOrSelf)
	rpeBuilder1.addUdpeTest(newElementTest("listitem", nil, nil))

	fpeBuilder2 := newFpeBuilder()
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("listitem", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("Text", nil, nil))
	fpeBuilder2.addAxis(child)
	fpeBuilder2.addUdpeTest(newElementTest("keyword", nil, nil))

	fpe1 := fpeBuilder1.end()
	_, rec1 := udpeGlobalTable.addFpe(fpe1)
	rec1.setNudpeRecord(nudpeRec)

	rpe1 := rpeBuilder1.end()
	_, rec2 := udpeGlobalTable.addRpe(rpe1)
	rec2.setNudpeRecord(nudpeRec)

	fpe2 := fpeBuilder2.end()
	_, rec3 := udpeGlobalTable.addFpe(fpe2)
	rec3.setNudpeRecord(nudpeRec)

}
