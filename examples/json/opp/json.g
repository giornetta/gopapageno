func ParserPreallocMem(inputSize int, numThreads int) {
}

%%

%axiom Document

%%

Document : Object
{
};

Object : LCURLY RCURLY
{
} | LCURLY Members RCURLY
{
}
;

Members : Pair
{
} | Members COMMA Pair
{
};

Pair : STRING COLON Value
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
} | Elements COMMA Value
{
};