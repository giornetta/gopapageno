import (
	"math"
)

var parserPools []*gopapageno.Pool[int64]

func ParserPreallocMem(inputSize int, numThreads int) {
	parserPools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(2)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		parserPools[i] = gopapageno.NewPool[int64](poolSizePerThread)

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
	newValue := parserPools[thread].Get()
	*newValue = *$1.Value.(*int64) + *$3.Value.(*int64)
	$$.Value = newValue
} | T
{
	$$.Value = $1.Value
};

T : NUMBER
{
	$$.Value = $1.Value
} | LPAR E RPAR
{
    $$.Value = $2.Value
};