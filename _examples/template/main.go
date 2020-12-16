// template
// This example demonstrates usage within an html/template, by using the
// HTML method to display it in a form.

package main

import (
	"html/template"
	"log"
	"os"

	"github.com/djangulo/gotcha"
)

func main() {
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		<form action="#">
		<label>Email
			<input type="text" placeholder="Your email here" />
		</label>
		{{.Captcha.HTML}}
		</form>
	</body>
</html>`

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(tpl)
	check(err)

	challenge, err := gotcha.Gen(gotcha.NewConfig("en"))

	data := struct {
		Title   string
		Items   []string
		Captcha gotcha.Challenge
	}{
		Title:   "My page",
		Captcha: challenge,
	}

	err = t.Execute(os.Stdout, data)
	check(err)

}
