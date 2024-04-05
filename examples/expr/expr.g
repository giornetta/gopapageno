import (
	"math"
)

<<<<<<<< HEAD:examples/expr/expr.g
var parserInt64Pools []*gopapageno.Pool[int64]

// parserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {
	parserInt64Pools = make([]*gopapageno.Pool[int64], numThreads)

========
var parserInt64Pools []*common.Pool[int64]

/*
parserPreallocMem initializes all the memory pools required by the semantic function of the parser.
*/
func parserPreallocMem(inputSize int, numThreads int) {
	parserInt64Pools = make([]*common.Pool[int64], numThreads)
	
>>>>>>>> 1cf3da6372fc7d7484c9dd9e3b09fa6cf31f0e80:examples/arithmetic/arith.g
	avgCharsPerNumber := float64(4)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
<<<<<<<< HEAD:examples/expr/expr.g
		parserInt64Pools[i] = gopapageno.NewPool[int64](poolSizePerThread)
========
		parserInt64Pools[i] = common.NewPool[int64](poolSizePerThread)
>>>>>>>> 1cf3da6372fc7d7484c9dd9e3b09fa6cf31f0e80:examples/arithmetic/arith.g
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
