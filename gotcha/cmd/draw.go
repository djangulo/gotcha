/*
Copyright © 2020 Denis Angulo <djal@tuta.io>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"strings"

	"github.com/djangulo/gotcha/draw"
	"github.com/spf13/cobra"
)

var (
	// drawCmd represents the draw command
	drawCmd = &cobra.Command{
		Use:   "draw",
		Short: "Draw some text onto an image.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("draw called")
		},
	}
	fuzzers    fuzzerArg
	thickness  int
	slope      randFloat64Arg
	noise      float64
	col1, col2 colorArg
	scale      float64
	rotation   int
)

func init() {

	drawCmd.Flags().VarP(
		&fuzzers,
		"fuzzers",
		"f",
		`Comma separated list of fuzzer functions to use. Accepts any combination of
[b]ands, [r]andom circles, [c]oncentric circles, random [l]ines. Default is a random
combination of two.`,
	)
	drawCmd.Flags().Var(&col1, "color", "Color to use for all fuzzer function.")
	drawCmd.Flags().Var(&col2, "fg-color", "Secondary color to use for Bands and ConcentricCircles fuzzers.")
	drawCmd.Flags().Float64Var(&noise, "noise", 0.5, "Noise value for RandomLines, ConcentricCircles and RandomCircles.")
	drawCmd.Flags().IntVar(&thickness, "thickness", 10, "Band thickness in Bands and Concentric circles.")

	drawCmd.Flags().Var(&slope, "slope", "Slope value for Bands. Randomized if unset")
	drawCmd.Flags().AddFlag(captchaCmd.Flags().Lookup("outfile"))
	drawCmd.Flags().IntVar(&rotation, "font-rotation", 0, "Degrees to rotate fonts by.")
	drawCmd.Flags().Float64Var(&scale, "font-scale", 1.0, "Scale fonts by this value.")
	drawCmd.Flags().Float64Var(&scale, "font-scale", 1.0, "Scale fonts by this value.")
	rootCmd.AddCommand(drawCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// drawCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// drawCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type fuzzerArg []draw.FuzzFactory

func (f *fuzzerArg) String() string {
	return fmt.Sprint(*f)
}

func (f *fuzzerArg) Set(value string) error {

	for _, fuzz := range strings.Split(value, ",") {
		fuzz = strings.ToLower(fuzz)
		switch fuzz {
		case "b":
			*f = append(*f, draw.Bands)
		case "r":
			*f = append(*f, draw.RandomCircles)
		case "l":
			*f = append(*f, draw.RandomLines)
		case "c":
			*f = append(*f, draw.ConcentricCircles)
		default:
			switch v := rand.Float64(); {
			case (v >= 0) && (v < 0.2):
				*f = []draw.FuzzFactory{draw.Bands, draw.RandomLines}
			case (v >= 0.2) && (v < 0.4):
				*f = []draw.FuzzFactory{draw.Bands, draw.ConcentricCircles}
			case (v >= 0.4) && (v < 0.6):
				*f = []draw.FuzzFactory{draw.RandomLines, draw.ConcentricCircles}
			case (v >= 0.6) && (v < 0.8):
				*f = []draw.FuzzFactory{draw.RandomLines, draw.RandomCircles}
			case (v >= 0.8):
				*f = []draw.FuzzFactory{draw.RandomCircles, draw.ConcentricCircles}
			}
			return nil
		}
	}
	return nil
}

func (f *fuzzerArg) Type() string {
	return "draw.FuzzFactory"
}

type colorArg color.RGBA

func (f *colorArg) String() string {
	return fmt.Sprint(*f)
}

func (f *colorArg) Set(value string) error {
	s := strings.SplitN(value, ",", 4)
	var r, g, b, a uint8
	for i, v := range s {
		vv, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return err
		}
		switch i {
		case 0:
			r = uint8(vv)
		case 1:
			g = uint8(vv)
		case 2:
			b = uint8(vv)
		case 3:
			a = uint8(vv)
		}
	}
	f.R = r
	f.G = g
	f.B = b
	f.A = a

	return nil
}

func (f *colorArg) Type() string {
	return "color.RGBA"
}

type randFloat64Arg float64

func (r *randFloat64Arg) String() string {
	return fmt.Sprint(*r)
}

func (r *randFloat64Arg) Set(value string) error {
	if value == "" {
		*r = randFloat64Arg(rand.Float64()*4 - 2)
		return nil
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	*r = randFloat64Arg(f)
	return nil
}

func (r *randFloat64Arg) Type() string {
	return "float64"
}
