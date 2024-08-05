%axiom S

%%

S : E
{
    $$.Value = $1.Value
};

E : (V TIMES OPERATOR)+ TIMES
{
    switch ruleType {
    case gopapageno.RuleCyclic:
        $$.Value = $1.Value.(int64)
    case gopapageno.RuleAppendRight:
        $$.Value = $$.Value.(int64) * 2
    case gopapageno.RuleAppendLeft:
        $$.Value = $$.Value.(int64) * 2
    case gopapageno.RuleCombine:
        $$.Value = $1.Value.(int64) + $4.Value.(int64)
    }
};

V : NUMBER
{
    $$.Value = $1.Value
};

%%