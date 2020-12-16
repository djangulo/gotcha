package draw

import (
	"context"
	"image"
	"image/color"
	"testing"
)

func BenchmarkBands(b *testing.B) {
	for _, bb := range []struct {
		name       string
		w, h       int
		col1, col2 color.Color
		thickness  int
		slope      float64
	}{
		{"random,100x100", 100, 100, nil, nil, 0, 0.0},
		{"random,200x200", 200, 200, nil, nil, 0, 0.0},
		{"random,300x300", 300, 300, nil, nil, 0, 0.0},
		{"random,400x400", 400, 400, nil, nil, 0, 0.0},
		{"fixed slope,100x100", 100, 100, nil, nil, 0, 2.0},
		{"fixed slope,200x200", 200, 200, nil, nil, 0, 2.0},
		{"fixed slope,300x300", 300, 300, nil, nil, 0, 2.0},
		{"fixed slope,400x400", 400, 400, nil, nil, 0, 2.0},
		{"fixed thickness,100x100", 100, 100, nil, nil, 20, 0.0},
		{"fixed thickness,200x200", 200, 200, nil, nil, 20, 0.0},
		{"fixed thickness,300x300", 300, 300, nil, nil, 20, 0.0},
		{"fixed thickness,400x400", 400, 400, nil, nil, 20, 0.0},
		{"fixed slope,thickness,100x100", 100, 100, nil, nil, 20, 2.0},
		{"fixed slope,thickness,200x200", 200, 200, nil, nil, 20, 2.0},
		{"fixed slope,thickness,300x300", 300, 300, nil, nil, 20, 2.0},
		{"fixed slope,thickness,400x400", 400, 400, nil, nil, 20, 2.0},
	} {
		img := image.NewRGBA(image.Rect(0, 0, bb.w, bb.h))
		ctx := context.Background()
		if bb.col1 != nil {
			ctx = context.WithValue(ctx, FuzzColor1CtxKey, bb.col1)
		}
		if bb.col2 != nil {
			ctx = context.WithValue(ctx, FuzzColor2CtxKey, bb.col2)
		}
		if bb.thickness != 0 {
			ctx = context.WithValue(ctx, FuzzBandThicknessCtxKey, bb.thickness)
		}
		if bb.slope-0.0 > 0.01 {
			ctx = context.WithValue(ctx, FuzzSlopeCtxKey, bb.slope)
		}
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Bands(ctx)(img)
			}
		})
		img = nil
		ctx = nil
	}
}

func BenchmarkAccurateBands(b *testing.B) {
	for _, bb := range []struct {
		name       string
		w, h       int
		col1, col2 color.Color
		thickness  int
		slope      float64
	}{
		{"random,100x100", 100, 100, nil, nil, 0, 0.0},
		{"random,200x200", 200, 200, nil, nil, 0, 0.0},
		{"random,300x300", 300, 300, nil, nil, 0, 0.0},
		{"random,400x400", 400, 400, nil, nil, 0, 0.0},
		{"fixed slope,100x100", 100, 100, nil, nil, 0, 2.0},
		{"fixed slope,200x200", 200, 200, nil, nil, 0, 2.0},
		{"fixed slope,300x300", 300, 300, nil, nil, 0, 2.0},
		{"fixed slope,400x400", 400, 400, nil, nil, 0, 2.0},
		{"fixed thickness,100x100", 100, 100, nil, nil, 20, 0.0},
		{"fixed thickness,200x200", 200, 200, nil, nil, 20, 0.0},
		{"fixed thickness,300x300", 300, 300, nil, nil, 20, 0.0},
		{"fixed thickness,400x400", 400, 400, nil, nil, 20, 0.0},
		{"fixed slope,thickness,100x100", 100, 100, nil, nil, 20, 2.0},
		{"fixed slope,thickness,200x200", 200, 200, nil, nil, 20, 2.0},
		{"fixed slope,thickness,300x300", 300, 300, nil, nil, 20, 2.0},
		{"fixed slope,thickness,400x400", 400, 400, nil, nil, 20, 2.0},
	} {
		img := image.NewRGBA(image.Rect(0, 0, bb.w, bb.h))
		ctx := context.Background()
		if bb.col1 != nil {
			ctx = context.WithValue(ctx, FuzzColor1CtxKey, bb.col1)
		}
		if bb.col2 != nil {
			ctx = context.WithValue(ctx, FuzzColor2CtxKey, bb.col2)
		}
		if bb.thickness != 0 {
			ctx = context.WithValue(ctx, FuzzBandThicknessCtxKey, bb.thickness)
		}
		if bb.slope-0.0 > 0.01 {
			ctx = context.WithValue(ctx, FuzzSlopeCtxKey, bb.slope)
		}
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				AccurateBands(ctx)(img)
			}
		})
		img = nil
		ctx = nil
	}
}

func BenchmarkRandomLines(b *testing.B) {
	for _, bb := range []struct {
		name  string
		w, h  int
		col1  color.Color
		noise float64
	}{
		{"default noise,100x100", 100, 100, nil, 0.5},
		{"default noise,200x200", 200, 200, nil, 0.5},
		{"default noise,300x300", 300, 300, nil, 0.5},
		{"default noise,400x400", 400, 400, nil, 0.5},
		{"max noise,100x100", 100, 100, nil, 1.0},
		{"max noise,200x200", 200, 200, nil, 1.0},
		{"max noise,300x300", 300, 300, nil, 1.0},
		{"max noise,400x400", 400, 400, nil, 1.0},
		{"tiny noise,100x100", 100, 100, nil, 0.01},
		{"tiny noise,200x200", 200, 200, nil, 0.01},
		{"tiny noise,300x300", 300, 300, nil, 0.01},
		{"tiny noise,400x400", 400, 400, nil, 0.01},
	} {
		img := image.NewRGBA(image.Rect(0, 0, bb.w, bb.h))
		ctx := context.Background()
		if bb.col1 != nil {
			ctx = context.WithValue(ctx, FuzzColor1CtxKey, bb.col1)
		}
		if bb.noise-0.0 > 0.01 {
			ctx = context.WithValue(ctx, FuzzNoiseCtxKey, bb.noise)
		}
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				RandomLines(ctx)(img)
			}
		})
		img = nil
		ctx = nil
	}
}

func BenchmarkRandomCircles(b *testing.B) {
	for _, bb := range []struct {
		name  string
		w, h  int
		col1  color.Color
		noise float64
	}{
		{"default noise,100x100", 100, 100, nil, 0.5},
		{"default noise,200x200", 200, 200, nil, 0.5},
		{"default noise,300x300", 300, 300, nil, 0.5},
		{"default noise,400x400", 400, 400, nil, 0.5},
		{"max noise,100x100", 100, 100, nil, 1.0},
		{"max noise,200x200", 200, 200, nil, 1.0},
		{"max noise,300x300", 300, 300, nil, 1.0},
		{"max noise,400x400", 400, 400, nil, 1.0},
		{"tiny noise,100x100", 100, 100, nil, 0.01},
		{"tiny noise,200x200", 200, 200, nil, 0.01},
		{"tiny noise,300x300", 300, 300, nil, 0.01},
		{"tiny noise,400x400", 400, 400, nil, 0.01},
	} {
		img := image.NewRGBA(image.Rect(0, 0, bb.w, bb.h))
		ctx := context.Background()
		if bb.col1 != nil {
			ctx = context.WithValue(ctx, FuzzColor1CtxKey, bb.col1)
		}
		if bb.noise-0.0 > 0.01 {
			ctx = context.WithValue(ctx, FuzzNoiseCtxKey, bb.noise)
		}
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				RandomCircles(ctx)(img)
			}
		})
		img = nil
		ctx = nil
	}
}

func BenchmarkConcentricCircles(b *testing.B) {
	for _, bb := range []struct {
		name  string
		w, h  int
		col1  color.Color
		noise float64
	}{
		{"default noise,100x100", 100, 100, nil, 0.5},
		{"default noise,200x200", 200, 200, nil, 0.5},
		{"default noise,300x300", 300, 300, nil, 0.5},
		{"default noise,400x400", 400, 400, nil, 0.5},
		{"max noise,100x100", 100, 100, nil, 1.0},
		{"max noise,200x200", 200, 200, nil, 1.0},
		{"max noise,300x300", 300, 300, nil, 1.0},
		{"max noise,400x400", 400, 400, nil, 1.0},
		{"tiny noise,100x100", 100, 100, nil, 0.01},
		{"tiny noise,200x200", 200, 200, nil, 0.01},
		{"tiny noise,300x300", 300, 300, nil, 0.01},
		{"tiny noise,400x400", 400, 400, nil, 0.01},
	} {
		img := image.NewRGBA(image.Rect(0, 0, bb.w, bb.h))
		ctx := context.Background()
		if bb.col1 != nil {
			ctx = context.WithValue(ctx, FuzzColor1CtxKey, bb.col1)
		}
		if bb.noise-0.0 > 0.01 {
			ctx = context.WithValue(ctx, FuzzNoiseCtxKey, bb.noise)
		}
		b.Run(bb.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ConcentricCircles(ctx)(img)
			}
		})
		img = nil
		ctx = nil
	}
}
