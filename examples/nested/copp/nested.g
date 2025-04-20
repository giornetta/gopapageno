%axiom S

%%

S : E
{
    $$.Value = $1.Value
};

E : (V TIMES OPERATOR)+ V
{
    v1 := $1.Value.(int64)
    v2 := $4.Value.(int64)

    if ruleFlags.Has(gopapageno.RuleAppend) || ruleFlags.Has(gopapageno.RuleCombine) {
        $$.Value = $$.Value.(int64) + v2
    } else {
        $$.Value = v1 + v2
    }
};

V : NUMBER
{
    $$.Value = $1.Value
};

%%