import (
	"math"
)

var parserInt64Pools []*gopapageno.Pool[int64]

// ParserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {
	parserInt64Pools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(4)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		parserInt64Pools[i] = gopapageno.NewPool[int64](poolSizePerThread)

	}
}
%%

%axiom S

%%

S : E
{
	$$.Value = $1.Value
};

E : E PLUS T
{
	newValue := parserInt64Pools[thread].Get()
	*newValue = *$1.Value.(*int64) + *$3.Value.(*int64)
	$$.Value = newValue
} | T
{
	$$.Value = $1.Value
};

T : T TIMES F
{
    newValue := parserInt64Pools[thread].Get()
    *newValue = *$1.Value.(*int64) * *$3.Value.(*int64)
    $$.Value = newValue
} | F
{
	$$.Value = $1.Value
};

F : LPAR E RPAR
{
	$$.Value = $2.Value
} | NUMBER
{
	$$.Value = $1.Value
};
