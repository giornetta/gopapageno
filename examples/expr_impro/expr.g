// ParserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {

}
%%

%axiom S

%%

S : E
{
	$$.Value = $1.Value
};

E : E PLUS T
{
	*$1.Value.(*int64) = *$1.Value.(*int64) + *$3.Value.(*int64)
	$$.Value = $1.Value
} | T
{
	$$.Value = $1.Value
};

T : T TIMES F
{
    *$1.Value.(*int64) = *$1.Value.(*int64) * *$3.Value.(*int64)
    $$.Value = $1.Value
} | F
{
	$$.Value = $1.Value
};

F : LPAR E RPAR
{
	$$.Value = $2.Value
} | NUMBER
{
	$$.Value = $1.Value
};
