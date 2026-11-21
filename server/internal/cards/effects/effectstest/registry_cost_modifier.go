package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// registerPanics runs Register and reports the panic message, or "" if
// it did not panic.
func RegisterPanics(spec effects.Spec) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg, _ = r.(string)
			if msg == "" {
				msg = "panicked with a non-string value"
			}
		}
	}()
	effects.Register(spec)
	return ""
}
