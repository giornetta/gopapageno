%cut \n

%%

DIGIT [0-9]
EQUAL =
PLUS \+
TIMES \*

%%

{DIGIT}+
{
    num, err := strconv.ParseInt(text, 10, 64)
    if err != nil {
        return gopapageno.LexErr
    }

    token.Type = NUMBER
    token.Value = num
}

{TIMES}
{
    token.Type = TIMES
}

({PLUS}{EQUAL})+
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
)