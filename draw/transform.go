package draw

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"math"

	"gonum.org/v1/gonum/mat"
)

func Rotate(ctx context.Context, img draw.Image, deg int) draw.Image {
	if deg%360 == 0 {
		return img
	}
	rotM := mat.NewDense(2, 2, []float64{
		math.Cos(degToRad(-deg)),
		-math.Sin(degToRad(-deg)),
		math.Sin(degToRad(-deg)),
		math.Cos(degToRad(-deg)),
	})

	// size needs to be adjusted for 45 < deg < 135 and 225 < deg < 315
	var wide = 1.0
	if (deg >= 55 && deg < 125) || (deg >= 235 && deg < 305) {
		wide = 1.3
	}

	canvas := image.NewRGBA(
		image.Rectangle{
			image.Point{int(math.Round(-float64(img.Bounds().Dx()) * wide / 2)), -img.Bounds().Dy() / 2},
			image.Point{int(math.Round(float64(img.Bounds().Dx()) * wide / 2)), img.Bounds().Dy() / 2},
		},
	)
	// canvas := image.NewRGBA(img.Bounds())
	// draw.Draw(canvas, canvas.Bounds(), m, b.Min.Add(p), draw.Src)

	var bgColor = color.RGBA{192, 192, 192, 255}
	if v, ok := ctx.Value(BackgroundColor).(color.RGBA); ok {
		bgColor = v
	}

	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			var pt mat.Dense

			// when rotating, we offset both x and y by -img.height/2 and -img.width/2,
			// and avoid unwanted translation
			pt.Mul(
				rotM,
				mat.NewDense(
					2,
					1,
					[]float64{float64(x - img.Bounds().Dx()/2), float64(y - img.Bounds().Dy()/2)}),
			)
			canvas.Set(int(pt.At(0, 0)), int(pt.At(1, 0)), img.At(x, y))
		}
	}
	return canvas
}

func Scale(ctx context.Context, img draw.Image, c float64) draw.Image {
	if math.Abs(c-1.0) < 0.001 {
		return img
	}
	if c < 0.0 {
		c = 0.0
	}
	var fillFactor int = 1
	if c > 1.0 {
		fillFactor = int(math.Round(c / 0.5))
	}
	scaled := image.NewRGBA(
		image.Rect(
			0,
			0,
			int(math.Round(float64(img.Bounds().Max.X)*c)),
			int(math.Round(float64(img.Bounds().Max.Y)*c)),
		),
	)
	var bgColor = color.RGBA{192, 192, 192, 255}
	if v, ok := ctx.Value(BackgroundColor).(color.RGBA); ok {
		bgColor = v
	}
	draw.Draw(scaled, scaled.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)
	sc := mat.NewDiagDense(3, []float64{c, c, 1.0})
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			var pt mat.Dense
			pt.Mul(sc, mat.NewDense(3, 1, []float64{float64(x), float64(y), 1.0}))
			for i := -fillFactor; i < fillFactor; i++ {
				scaled.Set(
					int(math.Round(pt.At(0, 0)))+i,
					int(math.Round(pt.At(1, 0)))+i,
					img.At(x, y),
				)
			}
		}
	}

	return scaled
}

func degToRad(deg int) float64 {
	return float64(deg%360) * math.Pi / 180.0
}
