package effectstest

import (
	"strings"
	"testing"
)

func MustPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("no panic; wanted one mentioning %q", want)
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, want) {
			t.Fatalf("panic = %v, want one mentioning %q", r, want)
		}
	}()
	fn()
}
