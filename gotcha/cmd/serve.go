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
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/djangulo/gotcha"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/acme/autocert"
)

var (
	// serveCmd represents the serve command
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Launches the gotcha server.",
		Long: `The gotcha server will listen for requests and serve captcha challenges to clients.
The server also provides a self-contained js module that can be used to
retrieve and validate captcha challenges.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("serve called")
			mux := http.NewServeMux()
			manager := gotcha.NewManager()
			mux.Handle(endpoint, manager.Router(context.Background()))

			var srv *http.Server
			if len(autotlsHosts) > 0 {
				certManager := autocert.Manager{
					Prompt:     autocert.AcceptTOS,
					HostPolicy: autocert.HostWhitelist(autotlsHosts...), //Your domain here
					Cache:      autocert.DirCache("certs"),              //Folder for storing certificates
				}
				srv = &http.Server{
					Addr: ":" + port,
					TLSConfig: &tls.Config{
						GetCertificate: certManager.GetCertificate,
					},
					Handler: mux,
				}
				log.Println("Listening on port :" + port)
				// TODO: create fallback handler to feed into certManager.HTTPHandler
				go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
				log.Fatal(srv.ListenAndServeTLS("", "")) //Key and cert are coming from Let's Encrypt
				return
			}

			http.ListenAndServe(":"+port, mux)

		},
	}

	//
	autotlsHosts []string
	tlsCert      string
	tlsKey       string
	port         string
	storageURL   string
	endpoint     string
	publicURL    string
)

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.
	serveCmd.Flags().StringSliceVar(
		&autotlsHosts,
		"autotls-hosts",
		nil,
		`Register the hosts passed for LetsEncrypt certificates, and enables automatic
TLS.`,
	)
	serveCmd.Flags().StringVar(&tlsCert, "tls-cert", "", "Path to TLS certificate to use.")
	serveCmd.Flags().StringVar(&tlsKey, "tls-key", "", "Path to TLS key to use.")
	serveCmd.Flags().StringVarP(&port, "port", "p", "9000", "Port to listen at.")
	serveCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "/gotcha", "Mountpoint for the server routes.")
	serveCmd.Flags().StringVar(
		&storageURL,
		"storage-url",
		fmt.Sprintf(
			"fs://./<endpoint>/&root=%s",
			filepath.Join(os.TempDir(), "gotcha-assets"),
		),
		`Captcha file storage. See https://github.com/djangulo/go-storage
for viable connection strings.`,
	)
	serveCmd.Flags().StringVarP(&publicURL, "public-url", "u", "", "Public URL for js files.")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
