// ParserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {
}

%%

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