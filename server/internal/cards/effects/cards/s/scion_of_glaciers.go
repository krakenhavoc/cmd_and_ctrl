package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scion of Glaciers — Creature — Elemental {2}{U}{U}, 2/5 (Khans of Tarkir):
//
//	"{U}: This creature gets +1/-1 until end of turn."
//
// The shared self-pump row (ThisGetsRow) with a negative toughness: it can be
// activated repeatedly, and the Scion dies as a state-based action when its
// toughness reaches 0 (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d49cef5-8f9c-46cd-aa5e-7861fc1dcaa4",
		Name:         "Scion of Glaciers",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{U}", 1, -1, "Scion of Glaciers")},
	})
}
