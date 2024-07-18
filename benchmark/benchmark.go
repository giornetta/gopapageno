package benchmark

import (
	"context"
	"github.com/giornetta/gopapageno"
	"os"
	"testing"
)

func RunExpect[T comparable](b *testing.B, r *gopapageno.Runner, filename string, expected T) {
	b.StopTimer()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		b.Fatalf("could not read source file: %v", err)
	}

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		result, err := r.Run(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}

		if *result.Value.(*T) != expected {
			b.Fatalf("expected %v, got %v", expected, *result.Value.(*T))
		}
	}
}

func Run(b *testing.B, r *gopapageno.Runner, filename string) {
	b.StopTimer()

	bytes, err := os.ReadFile(filename)
	if err != nil {
		b.Fatalf("could not read source file: %v", err)
	}

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := r.Run(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}
	}
}
