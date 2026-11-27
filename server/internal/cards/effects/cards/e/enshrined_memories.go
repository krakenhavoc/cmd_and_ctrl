package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enshrined Memories — Sorcery {X}{G} (Betrayers of Kamigawa, #5602):
//
//	"Reveal the top X cards of your library. Put all creature cards revealed
//	 this way into your hand and the rest on the bottom of your library in any
//	 order."
//
// Lair Delve's reveal-and-take with X cards and only creature cards
// (RevealTopXTakeMatching); the rest go to the bottom in an order you
// choose. X is the one chosen as the spell was cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b77fb85-a6d8-4f3e-8493-99ec5b5a6d67",
		Name:         "Enshrined Memories",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve:    effects.RevealTopXTakeMatching("Enshrined Memories", effects.Creature()),
	})
}
