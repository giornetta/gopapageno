package gopapageno_test

import (
	"context"
	"github.com/giornetta/gopapageno"
	"testing"
)

func TestToken_Size(t *testing.T) {
	// tree:
	// 0
	// |-- 1
	//     |-- 3
	//     |-- 4
	//         |-- 5
	// |-- 2
	tree := &gopapageno.Token{
		Type: gopapageno.TokenType(0),
		Child: &gopapageno.Token{
			Type: gopapageno.TokenType(1),
			Next: &gopapageno.Token{
				Type: gopapageno.TokenType(2),
			},
			Child: &gopapageno.Token{
				Type: gopapageno.TokenType(3),
				Next: &gopapageno.Token{
					Type: gopapageno.TokenType(4),
					Child: &gopapageno.Token{
						Type: gopapageno.TokenType(5),
					},
				},
			},
		},
	}

	if s := tree.Size(); s != 6 {
		t.Errorf("Expected size 6, got %v", s)
	}
}

func TestToken_Height(t *testing.T) {
	balancedTree := &gopapageno.Token{
		Type:       gopapageno.TokenType(0),
		Precedence: 0,
		Value:      nil,
		Next:       nil,
		Child: &gopapageno.Token{
			Type:       gopapageno.TokenType(1),
			Precedence: 0,
			Value:      nil,
			Next: &gopapageno.Token{
				Type:       gopapageno.TokenType(2),
				Precedence: 0,
				Value:      nil,
				Next:       nil,
				Child: &gopapageno.Token{
					Type:       gopapageno.TokenType(5),
					Precedence: 0,
					Value:      nil,
					Next: &gopapageno.Token{
						Type:       gopapageno.TokenType(6),
						Precedence: 0,
						Value:      nil,
						Next:       nil,
						Child:      nil,
					},
					Child: nil,
				},
			},
			Child: &gopapageno.Token{
				Type:       gopapageno.TokenType(3),
				Precedence: 0,
				Value:      nil,
				Next: &gopapageno.Token{
					Type:       gopapageno.TokenType(4),
					Precedence: 0,
					Value:      nil,
					Next:       nil,
					Child:      nil,
				},
				Child: nil,
			},
		},
	}

	if h, err := balancedTree.Height(context.Background()); h != 3 || err != nil {
		t.Errorf("Balanced Tree expected 3, got %d", h)
	}

	tree := &gopapageno.Token{
		Type: gopapageno.TokenType(0),
		Child: &gopapageno.Token{
			Type: gopapageno.TokenType(1),
			Next: &gopapageno.Token{
				Type: gopapageno.TokenType(2),
			},
			Child: &gopapageno.Token{
				Type: gopapageno.TokenType(3),
				Next: &gopapageno.Token{
					Type: gopapageno.TokenType(4),
					Child: &gopapageno.Token{
						Type: gopapageno.TokenType(5),
					},
				},
			},
		},
	}

	if h, err := tree.Height(context.Background()); h != 4 || err != nil {
		t.Errorf("Balanced Tree expected 4, got %d", h)
	}
}
