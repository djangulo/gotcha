
.PHONY: bindata
# Generate all the assets.go files using go-bindata.
bindata:
	go get -u github.com/go-bindata/go-bindata/...
	go-bindata -prefix draw/ -pkg draw -o draw/assets.go draw/static/...
	go-bindata -fs -pkg gotcha -o ./assets.go locale/... static/...

minify:
	go get -u github.com/tdewolff/minify/cmd/minify
	minify -o static/js/gotcha.tmpl.min.js static/js/gotcha.tmpl.js
	minify -o static/css/gotcha.min.css static/css/gotcha.css

mindata: minify bindata

VERSION?=v0.1.0
.PHONY: build
xgo:
	go get -u github.com/karalabe/xgo
	mkdir -p "./bin/${VERSION}"
	xgo -out "bin/${VERSION}/gotcha-${VERSION}" -ldflags='-w' --targets=windows-6.3/amd64,linux/amd64 ./cmd/gotcha

build: bindata xgo
