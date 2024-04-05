%cut \n

LPAR    \(
RPAR    \)
PLUS    \+
TIMES   \*
DIGIT   [0-9]
SPACE   [ \t]
NEWLINE [\r\n]

%%

{LPAR}
{
    token.Type = LPAR
}
{RPAR}
{
	token.Type = RPAR
}
{TIMES}
{
	token.Type = TIMES
}
{PLUS}
{
    token.Type = PLUS
}
{DIGIT}+
{
    num := lexerInt64Pools[thread].Get()
    var err error

	*num, err = strconv.ParseInt(text, 10, 64)
	if err != nil {
		return gopapageno.LexErr
	}

	token.Type = NUMBER
	token.Value = num
}
{SPACE}
{
	return gopapageno.LexSkip
}
{NEWLINE}
{
	return gopapageno.LexSkip
}

%%
import (
	"strconv"
	"math"
)

var lexerInt64Pools []*gopapageno.Pool[int64]

// LexerPreallocMem initializes all the memory pools required by the lexer.
func LexerPreallocMem(inputSize int, numThreads int) {
	lexerInt64Pools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(4)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		lexerInt64Pools[i] = gopapageno.NewPool[int64](poolSizePerThread)
	}
}