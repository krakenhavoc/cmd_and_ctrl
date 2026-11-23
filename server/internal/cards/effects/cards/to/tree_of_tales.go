package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tree of Tales — Artifact Land (EDHREC rank 1538):
//
//	"{T}: Add {G}."
//
// A Forest that is also an artifact — Great Furnace's shape in green:
// it counts for metalcraft, grows Storm-Kiln Artist, and dies to
// Vandalblast. The printed type line carries the artifact type; the
// only thing a Spec has to add is the mana ability the engine will
// not synthesise for a land without the basic supertype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b4aa971-b919-4750-8388-33d4f42c9280",
		Name:         "Tree of Tales",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
