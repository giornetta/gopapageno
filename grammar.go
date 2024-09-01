package gopapageno

import (
	"context"
	"io"
	"log"
)

var (
	discardLogger = log.New(io.Discard, "", 0)
)

type parseResult[S any] struct {
	threadNum int
	stack     *S
}

type ParserFunc func(rule uint16, ruleType RuleFlags, lhs *Token, rhs []*Token, thread int)

// A ReductionStrategy defines which kind of algorithm should be executed
// when collecting and running multiple parsing passes.
type ReductionStrategy uint8

const (
	// ReductionSweep will run a single serial pass after combining data from the first `n` parallel runs.
	ReductionSweep ReductionStrategy = iota
	// ReductionParallel will combine adjacent parsing results and recursively run `n-1` parallel runs until one stack remains.
	ReductionParallel
	// ReductionMixed will run a limited number of parallel passes, then combine the remaining inputs to perform a final serial pass.
	ReductionMixed
)

func (s ReductionStrategy) String() string {
	switch s {
	case ReductionSweep:
		return "sweep"
	case ReductionParallel:
		return "parallel"
	case ReductionMixed:
		return "mixed"
	default:
		return "unknown"
	}
}

type ParsingStrategy uint8

const (
	// OPP is Operator-Precedence Parsing. It is the original parsing ReductionStrategy.
	OPP ParsingStrategy = iota
	// AOPP is Associative Operator Precedence Parsing.
	AOPP
	// COPP is Cyclic Operator Precedence Parsing.
	COPP
)

func (s ParsingStrategy) String() string {
	switch s {
	case OPP:
		return "OPP"
	case AOPP:
		return "AOPP"
	case COPP:
		return "COPP"
	default:
		return "UNKNOWN"
	}
}

type Grammar struct {
	NumTerminals    uint16
	NumNonterminals uint16

	MaxRHSLength    int
	Rules           []Rule
	CompressedRules []uint16

	MaxPrefixLength    int
	Prefixes           [][]TokenType
	CompressedPrefixes []uint16

	PrecedenceMatrix          [][]Precedence
	BitPackedPrecedenceMatrix []uint64

	Func         ParserFunc
	PreambleFunc PreambleFunc

	ParsingStrategy ParsingStrategy
}

type Parser interface {
	Parse(ctx context.Context, tokensLists []*LOS[Token]) (*Token, error)
}

func (g *Grammar) Parser(src []byte, opts *RunOptions) Parser {
	switch g.ParsingStrategy {
	case OPP:
		return NewOPParser(g, src, opts)
	case AOPP:
		return NewAOPParser(g, src, opts)
	case COPP:
		return NewCOPParser(g, src, opts)
	default:
		panic("unknown parser strategy")
	}
}

func (g *Grammar) precedence(t1 TokenType, t2 TokenType) Precedence {
	v1 := t1.Value()
	v2 := t2.Value()

	flatElementPos := v1*g.NumTerminals + v2
	elem := g.BitPackedPrecedenceMatrix[flatElementPos/32]
	pos := uint((flatElementPos % 32) * 2)

	return Precedence((elem >> pos) & 0x3)
}

func (g *Grammar) findRuleMatch(rhs []TokenType) (TokenType, uint16) {
	var pos uint16

	for _, key := range rhs {
		//Skip the value and rule num for each node (except the last)
		pos += 2
		numIndices := g.CompressedRules[pos]
		if numIndices == 0 {
			return TokenEmpty, 0
		}

		pos++
		low := uint16(0)
		high := numIndices - 1
		startPos := pos
		foundNext := false

		for low <= high {
			indexPos := low + (high-low)/2
			pos = startPos + indexPos*2
			curKey := g.CompressedRules[pos]

			if uint16(key) < curKey {
				high = indexPos - 1
			} else if uint16(key) > curKey {
				low = indexPos + 1
			} else {
				pos = g.CompressedRules[pos+1]
				foundNext = true
				break
			}
		}
		if !foundNext {
			return TokenEmpty, 0
		}
	}

	return TokenType(g.CompressedRules[pos]), g.CompressedRules[pos+1]
}

func (g *Grammar) findPrefixMatch(rhs []TokenType) (TokenType, uint16) {
	var pos uint16

	for _, key := range rhs {
		//Skip the value and rule num for each node (except the last)
		pos += 2
		numIndices := g.CompressedPrefixes[pos]
		if numIndices == 0 {
			return TokenEmpty, 0
		}

		pos++
		low := uint16(0)
		high := numIndices - 1
		startPos := pos
		foundNext := false

		for low <= high {
			indexPos := low + (high-low)/2
			pos = startPos + indexPos*2
			curKey := g.CompressedPrefixes[pos]

			if uint16(key) < curKey {
				high = indexPos - 1
			} else if uint16(key) > curKey {
				low = indexPos + 1
			} else {
				pos = g.CompressedPrefixes[pos+1]
				foundNext = true
				break
			}
		}
		if !foundNext {
			return TokenEmpty, 0
		}
	}

	return TokenType(g.CompressedPrefixes[pos]), g.CompressedPrefixes[pos+1]
}

func collectResults[S any](results []*S, resultCh <-chan parseResult[S], errCh <-chan error, n int) error {
	completed := 0
	for completed < n {
		select {
		case result := <-resultCh:
			results[result.threadNum] = result.stack
			completed++
		case err := <-errCh:
			return err
		}
	}

	return nil
}
