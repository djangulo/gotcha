package gotcha

type CtxKey uint8

const (
	Language CtxKey = iota + 1
	Expiry
	NoGzip
	CaptchaCtxKey
)

const (
	createNew CtxKey = iota + 101
)
