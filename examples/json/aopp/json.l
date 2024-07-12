%cut ,\n

LCURLY {
RCURLY }
LSQUARE \[
RSQUARE \]
COMMA ,
COLON :

QUOTE "
CHAR [^"\n]

TRUE true
FALSE false

DIGIT [0-9]
DOT \.
MINUS \-

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

{DIGIT}+
{
    token.Type = NUMBER
}
{MINUS}{DIGIT}+
{
    token.Type = NUMBER
}
{DIGIT}+{DOT}{DIGIT}+
{
    token.Type = NUMBER
}
{MINUS}{DIGIT}+{DOT}{DIGIT}+
{
    token.Type = NUMBER
}

{TRUE}
{
    token.Type = BOOL
}

{FALSE}
{
    token.Type = BOOL
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
