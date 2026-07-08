package envflag

import "testing"

func TestTruthy(t *testing.T) {
	truthy := []string{"1", "true", "TRUE", "yes", "on", "anything", " 1 "}
	for _, v := range truthy {
		if !Truthy(v) {
			t.Errorf("Truthy(%q) = false, want true", v)
		}
	}
	falsy := []string{"", "0", "false", "FALSE", "no", "off", "  ", " Off "}
	for _, v := range falsy {
		if Truthy(v) {
			t.Errorf("Truthy(%q) = true, want false", v)
		}
	}
}
