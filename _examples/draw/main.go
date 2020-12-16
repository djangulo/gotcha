package main

import (
	"context"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/djangulo/gotcha"
	"github.com/djangulo/gotcha/draw"
)

func main() {
	rand.Seed(time.Now().Unix())
	ctx, _ := gotcha.AddToContext(context.Background(),
		draw.FuzzBandThicknessCtxKey, 20, // randomized if not set
		draw.FuzzColor1CtxKey, color.RGBA{128, 0, 0, 128},
		draw.FuzzColor2CtxKey, color.RGBA{0, 0, 128, 128},
		draw.ScaleBy, 1.5,
		draw.RandomRotation, true,
		draw.RotateBy, -25, // degrees
		draw.FuzzNoiseCtxKey, 1.0,
		draw.FuzzSlopeCtxKey, 1.0, // randomized if not set
	)
	// ctx = context.WithValue(ctx, draw.FontVariantCtxKey, draw.InconsolataInverted) // randomized if not set

	img := draw.Gen(
		"Candy shop!",
		draw.Inconsolata(ctx),
		draw.ConcentricCircles(ctx),
	)
	fh, _ := os.Create("candy-shop.png")
	defer fh.Close()
	png.Encode(fh, img)
}
