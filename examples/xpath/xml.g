import (
    "sync"
)

var reductionPool = &sync.Pool{
	New: func() interface{} {
		return new(xpath.Reduction)
	},
}

var parserElementsPools []*gopapageno.Pool[xpath.Element]

// ParserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {
    poolSizePerThread := 10000

    parserElementsPools = make([]*gopapageno.Pool[xpath.Element], numThreads)
    for i := 0; i < numThreads; i++ {
        parserElementsPools[i] = gopapageno.NewPool[xpath.Element](poolSizePerThread)
    }
}

%%

%axiom ELEM

%%

ELEM : ELEM OPENTAG ELEM CLOSETAG
{
    openTag := $2.Value.(xpath.OpenTagSemanticValue)
    closeTag := $4.Value.(xpath.CloseTagSemanticValue)

    element := parserElementsPools[thread].Get()
    element.SetFromExtremeTags(openTag, closeTag)

    generativeNonTerminal := $1.Value.(xpath.NonTerminal)
    wrappedNonTerminal := $3.Value.(xpath.NonTerminal)
    reducedNonTerminal := xpath.NewNonTerminal().SetNode(element).SetDirectChildAndInheritItsChildren(generativeNonTerminal)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, generativeNonTerminal, wrappedNonTerminal)
    reduction.Handle()
    reductionPool.Put(reduction)

    $$.Value = reducedNonTerminal
} | OPENTAG ELEM CLOSETAG
{
    openTag := $1.Value.(xpath.OpenTagSemanticValue)
    closeTag := $3.Value.(xpath.CloseTagSemanticValue)

    element := parserElementsPools[thread].Get()
    element.SetFromExtremeTags(openTag, closeTag)

    wrappedNonTerminal := $2.Value.(xpath.NonTerminal)
    reducedNonTerminal := xpath.NewNonTerminal().SetNode(element)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, nil, wrappedNonTerminal)
    reduction.Handle()
    reductionPool.Put(reduction)

    $$.Value = reducedNonTerminal
} | ELEM OPENTAG CLOSETAG
{
    openTag := $2.Value.(xpath.OpenTagSemanticValue)
    closeTag := $3.Value.(xpath.CloseTagSemanticValue)

    element := parserElementsPools[thread].Get()
    element.SetFromExtremeTags(openTag, closeTag)

    generativeNonTerminal := $1.Value.(xpath.NonTerminal)
    reducedNonTerminal := xpath.NewNonTerminal().SetNode(element).SetDirectChildAndInheritItsChildren(generativeNonTerminal)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, generativeNonTerminal, nil)
    reduction.Handle()
    reductionPool.Put(reduction)

    $$.Value = reducedNonTerminal
} | OPENTAG CLOSETAG
{
    openTag := $1.Value.(xpath.OpenTagSemanticValue)
    closeTag := $2.Value.(xpath.CloseTagSemanticValue)

    element := parserElementsPools[thread].Get()
    element.SetFromExtremeTags(openTag, closeTag)

    reducedNonTerminal := xpath.NewNonTerminal().SetNode(element)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, nil, nil)
    reduction.Handle()
    reductionPool.Put(reduction)

    $$.Value = reducedNonTerminal
} | ELEM OPENCLOSETAG
{
    openCloseTag := $2.Value.(xpath.OpenCloseTagSemanticValue)

    element := parserElementsPools[thread].Get()
    element.SetFromSingleTag(openCloseTag)

    generativeNonTerminal := $1.Value.(xpath.NonTerminal)
    reducedNonTerminal := xpath.NewNonTerminal().SetNode(element).SetDirectChildAndInheritItsChildren(generativeNonTerminal)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, generativeNonTerminal, nil)
    reduction.Handle()
    reductionPool.Put(reduction)

    $$.Value = reducedNonTerminal
} | OPENCLOSETAG
{
    openCloseTag := $1.Value.(xpath.OpenCloseTagSemanticValue)

    element := parserElementsPools[thread].Get()
    element.SetFromSingleTag(openCloseTag)

    reducedNonTerminal := xpath.NewNonTerminal().SetNode(element)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, nil, nil)
    reduction.Handle()
    reductionPool.Put(reduction)

    $$.Value = reducedNonTerminal
} | ELEM TEXT
{
    tsv := $2.Value.(xpath.TextSemanticValue)

    text := new(xpath.Text)
    text.SetFromText(tsv)

    generativeNonTerminal := $1.Value.(xpath.NonTerminal)

    reducedNonTerminal := xpath.NewNonTerminal().SetNode(text).SetDirectChildAndInheritItsChildren(generativeNonTerminal)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, generativeNonTerminal, nil)
    reduction.Handle()
    reductionPool.Put(reduction)

    $$.Value = reducedNonTerminal
} | TEXT
{
    tsv := $1.Value.(xpath.TextSemanticValue)

    text := new(xpath.Text)
    text.SetFromText(tsv)

    reducedNonTerminal := xpath.NewNonTerminal().SetNode(text)

    reduction := reductionPool.Get().(*xpath.Reduction)
    reduction.Setup(reducedNonTerminal, nil, nil)
    reduction.Handle()
    reductionPool.Put(reduction)


    $$.Value = reducedNonTerminal
};