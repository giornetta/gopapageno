%cut \n
%preamble LexerPreallocMem

DIGIT [0-9]
EQUAL =
PLUS \+

%%

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

{PLUS}({EQUAL}{EQUAL})+
{
    token.Type = OPERATOR
}

[\r\n \t]
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