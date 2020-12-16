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

	"github.com/spf13/cobra"
)

var (
	// captchaCmd represents the captcha command
	captchaCmd = &cobra.Command{
		Use:   "captcha",
		Short: "Create a one-off captcha image+audio pair.",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("captcha called")
		},
	}
	outfile  string
	wavfile  string
	stdout   bool
	language string
)

func init() {
	captchaCmd.Flags().BoolVar(
		&stdout,
		"stdout",
		false,
		"Output the base64 encoding to stdout, and speak the audio file.",
	)
	captchaCmd.Flags().StringVarP(
		&outfile,
		"outfile",
		"o",
		"captcha.png",
		"Filename to use for image file.",
	)
	captchaCmd.Flags().StringVarP(
		&wavfile,
		"wavfile",
		"w",
		"captcha.wav",
		"Filename to use for wav file.",
	)
	captchaCmd.Flags().StringVarP(
		&language,
		"lang",
		"l",
		"en",
		`Language for challenge and wav file. Note that default only includes en, es and
fr, other language symbols and question banks have to be included with --math-files or --bank-files setting.`,
	)
	rootCmd.AddCommand(captchaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// captchaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// captchaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// func randomString(n int) string {
// 	b := make([]byte, n)
// 	for i, cache, remain := n-1, rand.Int63(), max; i >= 0; {
// 		if remain == 0 {
// 			cache, remain = rand.Int63(), max
// 		}
// 		if idx := int(cache & mask); idx < len(chars) {
// 			b[i] = chars[idx]
// 			i--
// 		}
// 		cache >>= bitsNeeded
// 		remain--
// 	}

// 	return *(*string)(unsafe.Pointer(&b))
// }
