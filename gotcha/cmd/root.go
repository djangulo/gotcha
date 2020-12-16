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
	"os"
	"strings"

	"github.com/djangulo/gotcha"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	bankfiles mapArg
	mathfiles mapArg
	source    sources
	quiet     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gotcha",
	Short: "A versatile captcha generating solution.",
	Long: `gotcha provides tools to generate captcha images and audio, allowing
customization based on a Source.

Math source provides math problems in mixed text/numbers form:
	6 divided by three (acceptable answers: 2, two)
	seven × nine (acceptable answers: sixty-three, 63)
Random source provides an n-long string of random alphanumeric characters:
	1jiAoF (acceptable answers: 1jiAoF)
The QuestionBank source uses user-provided questions and answers, such as
	How many legs does a horse have? (acceptable answers: 4, four)
The QuestionBank, while being the most cumbersome, is the one source guaranteed
to not be broken by an advanced OCR script, given the turing-test like nature of
the questions.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gotcha.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().VarP(
		&source, "sources", "s",
		`Comma separated list of values for [m]ath, [b]ank
and [r]andom. Accepts any combination of [m,math],
[b,bank,question-bank,questions],[r,rand,random].`,
	)
	rootCmd.PersistentFlags().VarP(
		&bankfiles,
		"bank-files",
		"b",
		`Comma-separated list of lang:path/to/file.json
pairs containing questions to and acceptable
answers to use. The json file schema should be
	[{"question": "What is the first letter of the alphabet",
		"answers": ["ans1", "ans2"]},
		...]
See locale/en/bank.json for a full example.`,
	)
	rootCmd.PersistentFlags().VarP(
		&mathfiles,
		"math-files",
		"m",
		`Comma-separated list of lang:path/to/file.json pairs containing
math symbols to use. The json file schema should be
	[{"symbol": "1", "human": "one"}, ...]
and should include symbol entries for operators plus [+,U+002B],
minus [-, U+002D], times [×, U+00D7], and divided by [÷, U+00F7].
See locale/en/symbols.json for a full example.`,
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gotcha" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gotcha")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

type sources gotcha.Source

func (s *sources) String() string {
	return fmt.Sprint(*s)
}

func (s *sources) Set(value string) error {
	var x gotcha.Source
	for _, src := range strings.Split(value, ",") {
		src = strings.ToLower(src)
		switch src {
		case "m", "math":
			x |= gotcha.Math
		case "b", "bank", "question-bank", "questions":
			x |= gotcha.QuestionBank
		case "r", "rand", "random":
			x |= gotcha.Random
		default:
			x = gotcha.Math | gotcha.Random
		}
		*s = sources(x)
	}
	return nil
}

func (s *sources) Type() string {
	return "gotcha.Source"
}

type mapArg map[string]string

func (m *mapArg) String() string {
	return fmt.Sprint(*m)
}

// Set expects value to have the form lang:path,lang2:path2.
func (m *mapArg) Set(value string) error {
	if *m == nil {
		*m = make(mapArg)
	}
	for _, s := range strings.Split(value, ",") {
		ss := strings.SplitN(s, ":", 2)
		(*m)[ss[0]] = ss[1]
	}
	return nil
}

func (s *mapArg) Type() string {
	return "map[string]string"
}
