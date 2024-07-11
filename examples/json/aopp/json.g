%axiom Document

%%

Document : Object
{
};

Object : LCURLY RCURLY
{
} | LCURLY Members RCURLY
{
};

Members : Members COMMA Members
{
} | STRING COLON Value
{
};

Value : STRING
{
} | Array
{
} | Object
{
} | NUMBER
{
} | BOOL
{
};

Array : LSQUARE RSQUARE
{
} | LSQUARE Elements RSQUARE
{
};

Elements : Value
{
} | Elements COMMA Elements
{
};

%%

func ParserPreallocMem(inputSize int, numThreads int) {
}