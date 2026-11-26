package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trigon of Mending — Artifact {2} (Scars of Mirrodin, S82 OLD batch 470,
// #5577):
//
//	"This artifact enters with three charge counters on it.
//	 {W}{W}, {T}: Put a charge counter on this artifact.
//	 {2}, {T}, Remove a charge counter from this artifact: Target player
//	 gains 3 life."
//
// The Trigon shape (OldB470Trigon) with Tendrils-style life gain on the
// target player.
//
// No simplification.
func init() {
	effects.Register(effects.OldB470Trigon(
		effects.Spec{OracleID: "17b1c870-ef94-49af-8f9b-a8765f1b54db", Name: "Trigon of Mending", Completeness: effects.CompletenessFull},
		"{W}{W}", "Target player gains 3 life.",
		effects.TargetPlayer("target player"), game.Purpose{},
		func(g *game.Game, item *game.StackItem) error {
			return effects.TargetPlayerGainsLife(3)(item, effects.NewContext(g, item))
		}))
}
