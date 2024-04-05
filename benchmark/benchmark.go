package benchmark

import (
	"context"
	"github.com/giornetta/gopapageno"
	"os"
	"testing"
)

func Run[T comparable](b *testing.B, p *gopapageno.Parser, filename string, expected T) {
	b.StopTimer()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		b.Fatalf("could not read source file: %v", err)
	}

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		result, err := p.Parse(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}

		if *result.Value.(*T) != expected {
			b.Fatalf("expected %v, got %v", expected, *result.Value.(*T))
		}
	}
}
