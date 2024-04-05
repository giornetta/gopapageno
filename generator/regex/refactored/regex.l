%cut \n

LPAR \(
RPAR \)
SQLPAR \[
SQRPAR \]
STAR \*
PLUS \+
DASH \-
PIPE \|
CARET \^
ANY \.
SPACE   [ \t]
NEWLINE [\r\n]
ESCAPE \\

%%

{LPAR}
{
    token.Type = LPar
}
{RPAR}
{
	token.Type = RPar
}
{SQLPAR}
{
    insideSet = true
    token.Type = SquareLPar
}
{SQRPAR}
{
    insideSet = false
	token.Type = SquareRPar
}
{STAR}
{
	token.Type = Star
}
{PLUS}
{
    token.Type = Plus
}
{DASH}
{
	token.Type = Dash
}
{PIPE}
{
    token.Type = Pipe
}
{CARET}
{
	token.Type = Caret
}
{ANY}
{
    var anyCharClass [256]bool

    //Skip the first char (empty transition)
    for i := 1; i < len(anyCharClass); i++ {
        anyCharClass[i] = true
    }
    anyCharClass['\n'] = false
    anyCharClass['\r'] = false

    newNfa := newNfaFromCharClass(anyCharClass)

    token.Type = Any
    token.Value = &newNfa
}
{SPACE}
{
	return gopapageno.LexSkip
}
{NEWLINE}
{
	return gopapageno.LexSkip
}
{ESCAPE}[{LPAR}{RPAR}{SQLPAR}{SQRPAR}{STAR}{PLUS}{DASH}{PIPE}{CARET}{ANY}{ESCAPE}]
{
    token.Type = Char
    if insideSet {
        token.Type = CharInSet
    }

    token.Value = text[1]
}
.
{
    token.Type = Char
    if insideSet {
        token.Type = CharInSet
    }

    token.Value = text[0]
}

%%

var insideSet bool
