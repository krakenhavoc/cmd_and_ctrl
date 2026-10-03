package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phantoms.go — the Judgment Phantoms (ADR 0108 §8, #1906): "This
// creature enters with N +1/+1 counters on it. If damage would be dealt to
// this creature, prevent that damage. Remove a +1/+1 counter from this
// creature."
//
// The prevention is a static with an additional effect, one application
// per recipient in a damage instance: blocked by three creatures, all of
// the damage is prevented and one counter is removed; under damage that
// can't be prevented the damage is dealt and a counter is still removed
// (CR 615.12); with no counter left, the damage is still prevented (the
// rulings).

// phantomReplacements is a Phantom's two replacements: it enters with `n`
// +1/+1 counters, and its prevention static.
func phantomReplacements(name string, n int) []game.ReplacementEffect {
	return []game.ReplacementEffect{
		b10EntersWithCounters(game.CounterPlusOne, n, fmt.Sprintf("%s: enters with %d +1/+1 counters", name, n)),
		PreventDamageDealtTo(PreventionStatic{
			To:    ToThisCreature,
			Then:  removeACounterFromThisBody,
			Label: name + " — prevent damage to it and remove a +1/+1 counter",
		}),
	}
}
