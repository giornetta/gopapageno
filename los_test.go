package gopapageno

import (
	"testing"
)

func TestListOfStacks_Get(t *testing.T) {
	los := NewListOfStacks[Token](NewPool[stack[Token]](2))

	token := Token{
		Type:       TokenType(0),
		Precedence: PrecEmpty,
		Value:      42,
	}

	t1 := los.Push(token)

	if t1.Value != token.Value {
		t.Errorf("Expected %v, got %v", token.Value, t1.Value)
	}

	t2 := los.Get()

	if t1 != t2 {
		t.Errorf("Expected %v, got %v", t1, t2)
	}

	for i := range stackSize {
		los.Push(Token{
			Value: i,
		})
	}

	t3 := los.Get()
	if t3.Value.(int) != stackSize-1 {
		t.Errorf("Expected %v, got %v", stackSize-1, t3.Value.(int))
	}

	if los.cur == los.head {
		t.Errorf("Didn't go to second stack")
	}

	los.Pop()

	t4 := los.Get()
	if t4.Value.(int) != stackSize-2 {
		t.Errorf("Expected %v, got %v", stackSize-2, t4.Value.(int))
	}
}

func TestListOfTokenPointerStacks_Combine(t *testing.T) {
	list := newListOfTokenPointerStacks(NewPool[tokenPointerStack](1))

	list.Push(&Token{
		Type:       0,
		Precedence: PrecEmpty,
	})
	list.Push(&Token{
		Type:       1,
		Precedence: PrecEmpty,
	})
	// Last element of S^L (a, *)
	list.Push(&Token{
		Type:       2,
		Precedence: PrecTakes,
	})

	list.Push(&Token{
		Type:       3,
		Precedence: PrecYields,
	})
	list.Push(&Token{
		Type:       4,
		Precedence: PrecEquals,
	})
	// Last element of S^R (b, *)
	list.Push(&Token{
		Type:       5,
		Precedence: PrecYields,
	})
	list.Push(&Token{
		Type:       6,
		Precedence: PrecTakes,
	})

	combined := list.Combine()
	combinedIt := combined.HeadIterator()

	var lastTok *Token
	expected := 2
	for tok := combinedIt.Next(); tok != nil; tok = combinedIt.Next() {
		if tok.Type != TokenType(expected) {
			t.Errorf("Expected: %v, got %v", expected, tok.Value.(int))
		}
		expected++

		lastTok = tok
	}

	if lastTok == nil {
		t.Fatalf("Last token is nil.")
	}

	if lastTok.Type != TokenType(5) {
		t.Errorf("Expected: %v, got %v", 5, lastTok.Value.(int))
	}
}
