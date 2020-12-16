package draw

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math/rand"
	"strings"
	"sync"
)

var (
	inconsolataCursor map[rune]*image.Rectangle
	inconsolataFonts  map[rune]draw.Image
	inconsolataSrc    image.Image
)

type FontVariant uint8

const (
	// VariantInconsolataGray grayscale.
	VariantInconsolataGray FontVariant = iota + 1
	// VariantInconsolataBlack black font on white background.
	VariantInconsolataBlack
	// VariantInconsolataInverted white font on black background.
	VariantInconsolataInverted
)

func (v FontVariant) path() string {
	return [...]string{
		// VariantInconsolataGray:     "static/inconsolata/inconsolata_gray.png",
		// VariantInconsolataBlack:    "static/inconsolata/inconsolata.png",
		// VariantInconsolataInverted: "static/inconsolata/inconsolata_black.png",
		VariantInconsolataGray:     "inconsolata_gray.png",
		VariantInconsolataBlack:    "inconsolata.png",
		VariantInconsolataInverted: "inconsolata_black.png",
	}[v]
}

var lastVariant FontVariant

// Inconsolata draws text on img using the inconsolata font. ctx checks
//     - InconsolataVariant under FontVariantCtxKey key
func Inconsolata(ctx context.Context) func(string) draw.Image {
	const (
		chars   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`~!@#$%^&*(){}[]'\"<>,./=¿?+-_\\|;:‘’“”¡¢£€¥Š§š×÷‹›→←↑↓ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþÿĄĽŚŞŤŹŻŔĂĹĆČĘĚĎŃŇŐŘŮűłąăčďęľĺńňőřśşůŜĸŋΑαΒβΔδΕεΦφΓγΗηΙιΘθΚκΛλΜμΝνΟοΠπΧχΡρΣσΤτΥυΩωΞξΨψΖζ "
		xOffset = 24
		yOffset = 50
	)
	var variant = VariantInconsolataGray
	if v, ok := ctx.Value(FontVariantCtxKey).(FontVariant); ok {
		variant = v
	}
	var scale = 1.0
	if v, ok := ctx.Value(ScaleBy).(float64); ok {
		scale = v
	}
	var randRot = true
	var rotation = 0
	if v, ok := ctx.Value(RandomRotation).(bool); ok {
		randRot = v
	}
	if v, ok := ctx.Value(RotateBy).(int); ok {
		rotation = v
	}
	var lineBreak = 18
	if v, ok := ctx.Value(LineBreak).(int); ok {
		lineBreak = v
	}
	if inconsolataSrc == nil || lastVariant != variant {
		b64data := b64assets[variant.path()]
		b, err := base64.StdEncoding.DecodeString(b64data)
		// asset, err := Asset(variant.path())
		// if err != nil {
		// 	panic(err)
		// }
		fmt.Println(b64data)
		srcFH := bytes.NewReader(b)
		inconsolataSrc, _, err = image.Decode(srcFH)
		if err != nil {
			panic(err)
		}
		var x, y int
		inconsolataCursor = make(map[rune]*image.Rectangle)
		inconsolataFonts = make(map[rune]draw.Image)
		for _, r := range chars {
			sr := image.Rectangle{image.Point{x, y}, image.Point{x + xOffset, y + yOffset}}
			inconsolataCursor[r] = &sr
			fontImg := image.NewRGBA(
				image.Rectangle{
					image.Point{0, 0},
					image.Point{
						sr.Bounds().Max.X - sr.Bounds().Min.X,
						sr.Bounds().Max.Y - sr.Bounds().Min.Y,
					},
				})
			gray := color.RGBA{192, 192, 192, 255}
			draw.Draw(fontImg, fontImg.Bounds(), &image.Uniform{gray}, image.ZP, draw.Src)
			draw.Draw(fontImg, fontImg.Bounds(), inconsolataSrc, sr.Min, draw.Src)
			inconsolataFonts[r] = fontImg
			if r == 'z' ||
				r == 'Z' ||
				r == ']' ||
				r == '€' ||
				r == 'Í' ||
				r == 'è' ||
				r == 'Ş' ||
				r == 'ń' ||
				r == 'Θ' ||
				r == 'Ξ' {
				y += yOffset
				x = 0
			} else {
				x += xOffset
			}
		}
	}

	return func(text string) draw.Image {
		lines, longest := breakLines(text, lineBreak)
		var wide = 1.0
		if !randRot && (rotation >= 55 && rotation < 125) || (rotation >= 235 && rotation < 305) {
			wide = 1.3
		}

		img := image.NewRGBA(
			image.Rect(
				0,
				0,
				int(float64(xOffset)*float64(longest)*scale*wide),
				int(float64(yOffset)*float64(len(lines))*scale),
			),
		)
		var c color.RGBA
		switch variant {
		case VariantInconsolataBlack:
			ctx = context.WithValue(ctx, BackgroundColor, color.RGBA{255, 255, 255, 255})
		case VariantInconsolataInverted:
			ctx = context.WithValue(ctx, BackgroundColor, color.RGBA{0, 0, 0, 255})
		default:
			ctx = context.WithValue(ctx, BackgroundColor, color.RGBA{192, 192, 192, 255})
		}
		draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)

		var wg sync.WaitGroup
		xpos, ypos := 0, 0
		for _, line := range lines {
			// get font positions
			wg.Add(1)
			go func(line string, xpos, ypos int) {
				defer wg.Done()
				for _, run := range line {
					if glyph, ok := inconsolataFonts[run]; ok {
						if randRot {
							// -35 to 35 deg
							rotation = rand.Intn(75) - 35
						}
						scaled := Scale(ctx, glyph, scale)
						rotated := Rotate(ctx, scaled, rotation)
						// rotated := Rotate(scaled, rotation)
						r := image.Rectangle{
							image.Pt(xpos, ypos),
							image.Pt(xpos, ypos).Add(rotated.Bounds().Size()),
						}
						draw.Draw(
							img,
							r,
							rotated,
							rotated.Bounds().Min,
							draw.Over)
						xpos += rotated.Bounds().Dx()
						continue
					} else if sr, ok := inconsolataCursor[run]; ok {
						r := image.Rectangle{
							image.Pt(xpos, ypos),
							image.Pt(xpos, ypos).Add(sr.Size()),
						}
						// fontImg := image.NewRGBA(r)
						// // gray := color.RGBA{192, 192, 192, 255}
						// // draw.Draw(fontImg, fontImg.Bounds(), &image.Uniform{gray}, image.ZP, draw.Src)
						// // draw.Draw(fontImg, r, inconsolataSrc, sr.Min, draw.Src)
						// fimg := Scale(fontImg, 2.0)
						// draw.Draw(img, r, fimg, sr.Min, draw.Src)
						draw.Draw(img, r, inconsolataSrc, sr.Min, draw.Src)
						xpos += int(float64(xOffset) * scale)
						continue
					}
					log.Printf("unknown character %c (%[1]U)", run)
				}
			}(line, xpos, ypos)
			ypos += yOffset
			xpos = 0
		}
		wg.Wait()
		return img
	}

}

// breakLines breaks text into lines, at the first space at or after
// count characters.
func breakLines(text string, count int) (lines []string, longest int) {
	text = strings.ReplaceAll(text, "\t", "  ")
	lines = make([]string, 0)
	i, j := 0, 1
	for j < len(text) {
		if (text[j] == ' ' && len(text[i:j]) > count) || text[j] == '\n' {
			part := strings.TrimSpace(text[i:j])
			lines = append(lines, part)
			if len(part) > longest {
				longest = len(part)
			}
			i = j
			// account for the last bit
			if len(text[i:]) < count {
				lines = append(lines, strings.TrimSpace(text[i:]))
			}
		}
		j++
	}
	// if there are no linebreaks, set the longest
	// to the length of the text
	if len(lines) <= 1 {
		lines = []string{text}
		longest = len(text)
	}
	return
}
