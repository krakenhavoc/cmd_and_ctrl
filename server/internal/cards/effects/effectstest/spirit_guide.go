package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// spirit_guide_test.go — #1228's catalog half: the constructor, the
// two proof cards, and the three boot-time refusals that keep a
// half-declared hand mana ability from registering and then doing
// nothing.

// specByName finds a registered Spec by its printed name. The two
// proof cards below are keyed by oracle ID like every other card, and
// the name is what a reader of this file recognises.
func SpecByName(t *testing.T, name string) effects.Spec {
	t.Helper()
	for _, spec := range effects.All() {
		if spec.Name == name {
			return spec
		}
	}
	t.Fatalf("%s is not registered", name)
	return effects.Spec{}
}
