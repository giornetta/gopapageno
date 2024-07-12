%axiom ELEM

%%

ELEM : ELEM OpenBracket ELEM CloseBracket
{
} | ELEM OpenParams ELEM CloseParams
{
} | ELEM OpenCloseInfo
{
} | ELEM OpenCloseParams
{
} | ELEM AlternativeClose
{
} | OpenBracket ELEM CloseBracket
{
} | OpenParams ELEM CloseBracket
{
} | OpenCloseInfo
{
} | OpenCloseParams
{
} | AlternativeClose
{
} | Infos
{
};

%%
