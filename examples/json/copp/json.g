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
};

Members : (Pair COMMA)+ Pair
{
} | Pair
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
};

Array : LSQUARE RSQUARE
{
} | LSQUARE Elements RSQUARE
{
};

Elements : (Value COMMA)+ Value
{
} | Value
{
};