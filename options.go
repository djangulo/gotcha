package gotcha

import (
	"time"

	gostorage "github.com/djangulo/go-storage"
)

type Option func(*Manager)

// WithSources sets the source to s.
func WithSources(s Source) Option {
	return func(m *Manager) {
		m.Sources = s
	}
}

// WithMountPoint sets the mountpoint to v.
func WithMountPoint(v string) Option {
	return func(m *Manager) {
		m.mountpoint = v
	}
}

// WithNoGzip serve the static assets without gzipping them.
func WithNoGzip() Option {
	return func(m *Manager) {
		m.noGzip = true
	}
}

// WithExpiry sets captcha expiration to expiry.
func WithExpiry(expiry time.Duration) Option {
	return func(m *Manager) {
		m.defaultExpiry = expiry
	}
}

// WithLifetime sets lifetime of captchas, once passed, to lifetime.
func WithLifetime(lifetime time.Duration) Option {
	return func(m *Manager) {
		m.lifetimeAfterPassed = lifetime
	}
}

// WithMath appends Math to the sources and replaces the default math symbols
// with the passed symbols.
func WithMath(symbols *Symbols) Option {
	return func(m *Manager) {
		m.Sources |= Math
		m.Math = symbols
	}
}

// WithStorage sets the Store to the storageURL.
// See https://github.com/djangulo/go-storage for viable connection strings.
// Panics on error.
func WithStorage(storageURL string) Option {
	return func(m *Manager) {
		storage, err := gostorage.Open(storageURL)
		if err != nil {
			panic(err)
		}
		m.FileStorage = storage
	}
}

func WithQuestionBank(bank *Bank) Option {
	return func(m *Manager) {
		m.Sources |= QuestionBank
		m.Bank = bank
	}
}
