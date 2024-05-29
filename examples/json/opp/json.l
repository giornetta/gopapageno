%cut ,

LCURLY {
RCURLY }
LSQUARE \[
RSQUARE \]
COMMA ,
COLON :
QUOTE "
CHAR [!#-\[\]-ÿ]
SPACE [ \t]
NEWLINE [\r\n]

%%

{LCURLY}
{
    token.Type = LCURLY
}
{RCURLY}
{
    token.Type = RCURLY
}

{LSQUARE}
{
    token.Type = LSQUARE
}
{RSQUARE}
{
    token.Type = RSQUARE
}

{COMMA}
{
    token.Type = COMMA
}
{COLON}
{
    token.Type = COLON
}

{QUOTE}{CHAR}*{QUOTE}
{
    token.Type = STRING
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

func LexerPreallocMem(inputSize int, numThreads int) {
}