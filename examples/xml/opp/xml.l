%cut <

LBRACKET <
RBRACKET >
LSLASH   </
RSLASH   />
EQUALS   =
VALUE    ".*"
NUMBER   [0-9]+
IDENT    [a-zA-Z0-9_\-\."=][a-zA-Z0-9_\-\."= ~]*
INFOS    [a-zA-Z0-9_:#/,&;\.=\-\(\)@?\[\]\+'$"!][a-zA-Z0-9_:#/,&;\.=\-\(\)@?\[\]\+'$"! ~]*
SPACE    [ \t]+
NEWLINE  [\r\n]

%%

{INFOS}
{
	token.Type = Infos
}
{LBRACKET}{IDENT}{RBRACKET}
{
	token.Type = OpenBracket
}
{LSLASH}{IDENT}{RBRACKET}
{
	token.Type = CloseBracket
}
{LBRACKET}{IDENT}{RSLASH}
{
	token.Type = AlternativeClose
}
{LBRACKET}{IDENT}{SPACE}{IDENT}{EQUALS}{VALUE}{RBRACKET}
{
    token.Type = OpenParams
}
{LBRACKET}{IDENT}{RBRACKET}{LSLASH}{IDENT}{RBRACKET}
{
	token.Type = OpenCloseInfo
}
{LBRACKET}{IDENT}{SPACE}{IDENT}{EQUALS}{VALUE}{RBRACKET}{LSLASH}{IDENT}{RBRACKET}
{
	token.Type = OpenCloseParams
}
{SPACE}
{
	return gopapageno.LexSkip
}
{NEWLINE}
{
	return gopapageno.LexSkip
}
<?[^?]+?>
{
	return gopapageno.LexSkip
}

%%
