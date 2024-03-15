package gopapageno

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"runtime"
	"runtime/pprof"
	"time"
)

type Parser interface {
	Parse(r io.Reader) (*Symbol, error)
}

/*
symbol contains a token and its value, a precedence and pointers to build the syntactic tree.
*/
type Symbol struct {
	Token      uint16
	Precedence uint16
	Value      interface{}
	Next       *Symbol
	Child      *Symbol
}

/*
ParseString parses a string in parallel using an operator precedence grammar.
It takes as input a string as a slice of bytes and the number of threads, and returns a boolean
representing the success or failure of the parsing and the symbol at the root of the syntactic tree (if successful).
*/
func Parse(r io.Reader, numThreads int) (*Symbol, error) {
	rawInputSize := len(str)
	r.
		avgCharsPerToken := float64(12.5)

	//The last multiplication by  is to account for the generated nonterminals
	stackPoolBaseSize := math.Ceil((((float64(rawInputSize) / avgCharsPerToken) / float64(_STACK_SIZE)) / float64(numThreads)))
	stackPtrPoolBaseSize := math.Ceil(((float64(rawInputSize) / avgCharsPerToken) / float64(_STACK_PTR_SIZE)) / float64(numThreads))

	//Stats.StackPoolSize = stackPoolSize
	//Stats.StackPtrPoolSize = stackPtrPoolSize

	//Alloc memory required by both the lexer and the parser.
	//The call to runtime.GC() avoids the garbage collector to run concurrently with the parser
	start := time.Now()

	stackPools := make([]*stackPool, numThreads)
	stackPoolsNewNonterminals := make([]*stackPool, numThreads)
	stackPtrPools := make([]*stackPtrPool, numThreads)

	Stats.StackPoolSizes = make([]int, numThreads)
	Stats.StackPoolNewNonterminalsSizes = make([]int, numThreads)
	Stats.StackPtrPoolSizes = make([]int, numThreads)

	for i := 0; i < numThreads; i++ {
		stackPools[i] = newStackPool(int(stackPoolBaseSize * 1.2))
		Stats.StackPoolSizes[i] = int(stackPoolBaseSize * 1.2)
		stackPoolsNewNonterminals[i] = newStackPool(int(stackPoolBaseSize * 0.8))
		Stats.StackPoolNewNonterminalsSizes[i] = int(stackPoolBaseSize * 0.8)
		stackPtrPools[i] = newStackPtrPool(int(stackPtrPoolBaseSize))
		Stats.StackPtrPoolSizes[i] = int(stackPtrPoolBaseSize)
	}

	var stackPoolFinalPass *stackPool
	var stackPoolNewNonterminalsFinalPass *stackPool
	var stackPtrPoolFinalPass *stackPtrPool
	if numThreads > 1 {
		stackPoolFinalPass = newStackPool(int(math.Ceil(stackPoolBaseSize * 0.1 * float64(numThreads))))
		Stats.StackPoolSizeFinalPass = int(math.Ceil(stackPoolBaseSize * 0.1 * float64(numThreads)))
		stackPoolNewNonterminalsFinalPass = newStackPool(int(math.Ceil(stackPoolBaseSize * 0.05 * float64(numThreads))))
		Stats.StackPoolNewNonterminalsSizeFinalPass = int(math.Ceil(stackPoolBaseSize * 0.05 * float64(numThreads)))
		stackPtrPoolFinalPass = newStackPtrPool(int(math.Ceil(stackPtrPoolBaseSize * 0.1)))
		Stats.StackPtrPoolSizeFinalPass = int(int(math.Ceil(stackPtrPoolBaseSize * 0.1)))
	}

	lexerPreallocMem(rawInputSize, numThreads)

	parserPreallocMem(rawInputSize, numThreads)

	runtime.GC()

	Stats.AllocMemTime = time.Since(start)

	//Lex the file to obtain the input list
	start = time.Now()

	cutPoints, numLexThreads := findCutPoints(str, numThreads)

	Stats.NumLexThreads = numLexThreads
	Stats.LexTimes = make([]time.Duration, numLexThreads)
	Stats.CutPoints = cutPoints

	if numLexThreads < numThreads {
		fmt.Printf("It was not possible to find cut points for all %d threads.\n", numThreads)
		fmt.Printf("The number of lexing threads was reduced to %d.\n", numLexThreads)
	}

	lexC := make(chan lexResult)

	for i := 0; i < numLexThreads; i++ {
		go lex(i, str[cutPoints[i]:cutPoints[i+1]], stackPools[i], lexC)
	}

	lexResults := make([]lexResult, numLexThreads)

	for i := 0; i < numLexThreads; i++ {
		curLexResult := <-lexC
		lexResults[curLexResult.threadNum] = curLexResult

		if !curLexResult.success {
			Stats.LexTimeTotal = time.Since(start)
			return nil, errors.New("Lexing error")
		}
	}

	input := lexResults[0].tokenList

	for i := 1; i < numLexThreads; i++ {
		input.Merge(*lexResults[i].tokenList)
	}

	//input, err := lex(str, stackPool, lexC)

	Stats.LexTimeTotal = time.Since(start)

	//If lexing fails, abort the parsing
	/*if err != nil {
		fmt.Println(err.Error())
		return false, nil
	}*/

	if cpuprofileFile != nil {
		if err := pprof.StartCPUProfile(cpuprofileFile); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	Stats.NumTokensTotal = input.Length()

	if input.Length() == 0 {
		return nil, nil
	}

	start = time.Now()

	var result *symbol = nil

	//If there are not enough stacks in the input, reduce the number of threads.
	//This is because the input is split by splitting stacks, not stack contents.
	if input.NumStacks() < numThreads {
		fmt.Println("There are less stacks than threads, reducing the number of threads to", input.NumStacks())
		numThreads = input.NumStacks()
	}

	Stats.NumParseThreads = numThreads
	Stats.ParseTimes = make([]time.Duration, numThreads)

	//Split the input list
	inputLists := input.Split(numThreads)

	Stats.NumTokens = make([]int, numThreads)
	for i := 0; i < numThreads; i++ {
		Stats.NumTokens[i] = inputLists[i].Length()
	}

	parseResults := make([]parseResult, numThreads)

	c := make(chan parseResult)

	//Create the thread contexts and run the threads
	for i := 0; i < numThreads; i++ {
		//fmt.Print("Thread", i, " input: ")
		//threadContexts[i].input.Println()
		//fmt.Print("Thread", i, " stack: ")
		//threadContexts[i].stack.Println()

		var nextSym *symbol = nil

		if i < numThreads-1 {
			nextInputListIter := inputLists[i+1].HeadIterator()
			nextSym = nextInputListIter.Next()
		}

		go threadJob(numThreads, i, false, &inputLists[i], nextSym, stackPoolsNewNonterminals[i], stackPtrPools[i], c)

		/*threadContexts[i] = <-c

		fmt.Println("Thread", threadContexts[i].num, "finished parsing")
		fmt.Println("Result:", threadContexts[i].result)
		fmt.Print("Partial stack: ")
		threadContexts[i].stack.Println()

		if threadContexts[i].result == "failure" {
			fmt.Printf("Time to parse it: %s\n", time.Since(start))
			return false
		}*/
	}

	//Wait for each thread to finish its job
	for i := 0; i < numThreads; i++ {
		curParseResults := <-c

		parseResults[curParseResults.threadNum] = curParseResults

		//fmt.Println("Thread", threadContext.num, "finished parsing")
		//fmt.Println("Result:", threadContext.result)
		//fmt.Print("Partial stack: ")
		//threadContext.stack.Println()

		//If one of the threads fails, abort the parsing
		if !curParseResults.success {
			Stats.ParseTimeTotal = time.Since(start)
			return nil, errors.New("Parsing error")
		}
	}

	//Stats.RemainingStacks = stackPool.Remainder()
	//Stats.RemainingStackPtrs = stackPtrPool.Remainder()

	//If the number of threads is greater than one, a final pass is required
	if numThreads > 1 {
		startRecombiningStacks := time.Now()
		//Create the final input by joining together the stacks from the previous step
		finalPassInput := newLos(stackPoolFinalPass)
		for i := 0; i < numThreads; i++ {
			iterator := parseResults[i].stack.HeadIterator()
			//Ignore the first token
			iterator.Next()
			sym := iterator.Next()
			for sym != nil {
				finalPassInput.Push(sym)
				sym = iterator.Next()
			}
		}
		Stats.RecombiningStacksTime = time.Since(startRecombiningStacks)

		//fmt.Print("Final pass thread input: ")
		//finalPassThreadContext.input.Println()
		//fmt.Print("Final pass thread stack: ")
		//finalPassThreadContext.stack.Println()

		go threadJob(1, 0, true, &finalPassInput, nil, stackPoolNewNonterminalsFinalPass, stackPtrPoolFinalPass, c)

		finalPassParseResult := <-c

		//fmt.Println("Final thread finished parsing")
		//fmt.Println("Result:", finalPassThreadContext.result)
		//fmt.Print("Final stack: ")
		//finalPassThreadContext.stack.Println()

		if !finalPassParseResult.success {
			Stats.ParseTimeTotal = time.Since(start)
			return nil, errors.New("Parsing error")
		}

		//Pop tokens from the stack until a nonterminal is found
		sym := finalPassParseResult.stack.Pop()

		for isTerminal(sym.Token) {
			sym = finalPassParseResult.stack.Pop()
		}

		//sym.PrintTreeln()

		//Set the result as the nonterminal symbol
		result = sym

		Stats.RemainingStacksFinalPass = stackPoolFinalPass.Remainder()
		Stats.RemainingStacksNewNonterminalsFinalPass = stackPoolNewNonterminalsFinalPass.Remainder()
		Stats.RemainingStackPtrsFinalPass = stackPtrPoolFinalPass.Remainder()
	} else {
		//Pop tokens from the stack until a nonterminal is found
		sym := parseResults[0].stack.Pop()

		for isTerminal(sym.Token) {
			sym = parseResults[0].stack.Pop()
		}

		//sym.PrintTreeln()

		//Set the result as the nonterminal symbol
		result = sym
	}

	Stats.RemainingStacks = make([]int, numThreads)
	Stats.RemainingStacksNewNonterminals = make([]int, numThreads)
	Stats.RemainingStackPtrs = make([]int, numThreads)

	for i := 0; i < numThreads; i++ {
		Stats.RemainingStacks[i] = stackPools[i].Remainder()
		Stats.RemainingStacksNewNonterminals[i] = stackPoolsNewNonterminals[i].Remainder()
		Stats.RemainingStackPtrs[i] = stackPtrPools[i].Remainder()
	}

	Stats.ParseTimeTotal = time.Since(start)

	//Stats.RemainingStacks = stackPool.Remainder()
	//Stats.RemainingStackPtrs = stackPtrPool.Remainder()

	return result, nil
}

/*
ParseFile parses a file in parallel using an operator precedence grammar.
It takes as input a filename and the number of threads, and returns a boolean
representing the success or failure of the parsing and the symbol at the root of the syntactic tree (if successful).
*/
func ParseFile(filename string, numThreads int) (*symbol, error) {
	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return ParseString(bytes, numThreads)
}
