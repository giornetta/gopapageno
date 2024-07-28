package gopapageno

import (
	"testing"
)

func TestLOS_Get(t *testing.T) {
	los := NewLOS[Token](NewPool(2, WithConstructor(newStack[Token])))
	size := stackSize[Token]()

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

	for i := range size {
		los.Push(Token{
			Value: i,
		})
	}

	t3 := los.Get()
	if t3.Value.(int) != size-1 {
		t.Errorf("Expected %v, got %v", size-1, t3.Value.(int))
	}

	if los.cur == los.head {
		t.Errorf("Didn't go to second stack")
	}

	los.Pop()

	t4 := los.Get()
	if t4.Value.(int) != size-2 {
		t.Errorf("Expected %v, got %v", size-2, t4.Value.(int))
	}
}
