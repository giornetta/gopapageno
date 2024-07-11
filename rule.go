package gopapageno

type RuleType uint8

const (
	RuleSimple = iota
	RuleAppendLeft
	RuleAppendRight
	RuleCombine
	RuleCyclic
	RulePrefix
)

func (t RuleType) String() string {
	switch t {
	case RuleSimple:
		return "RuleSimple"
	case RuleAppendLeft:
		return "RuleAppendLeft"
	case RuleAppendRight:
		return "RuleAppendRight"
	case RuleCombine:
		return "RuleCombine"
	case RuleCyclic:
		return "RuleCyclic"
	default:
		return "Unknown"
	}
}

type Rule struct {
	Lhs  TokenType
	Rhs  []TokenType
	Type RuleType
}
