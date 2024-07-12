%cut \n
%preamble LexerPreallocMem

LPAR \(
RPAR \)
PLUS \+
DIGIT [0-9]
SPACE [ \t]
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
{PLUS}
{
    token.Type = PLUS
}
{DIGIT}+
{
    num := lexerPools[thread].Get()
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

var lexerPools []*gopapageno.Pool[int64]

func LexerPreallocMem(inputSize int, numThreads int) {
	lexerPools = make([]*gopapageno.Pool[int64], numThreads)

	avgCharsPerNumber := float64(2)
	poolSizePerThread := int(math.Ceil((float64(inputSize) / avgCharsPerNumber) / float64(numThreads)))

	for i := 0; i < numThreads; i++ {
		lexerPools[i] = gopapageno.NewPool[int64](poolSizePerThread)
	}
}