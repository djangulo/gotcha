package gotcha

import "testing"

func TestparseExpr(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want string
	}{
		{"2 × 8", "16"},
		{"2 + 1", "3"},
		{"16 + 16", "32"},
		{"100 × 100", "10000"},
		{"100 ÷ 2", "50"},
		{"99 ÷ 2", "49"},
	} {
		got, _ := parseExpr(tt.in)
		if got != tt.want {
			t.Errorf("expected %q got %q", tt.want, got)
		}
	}
}
