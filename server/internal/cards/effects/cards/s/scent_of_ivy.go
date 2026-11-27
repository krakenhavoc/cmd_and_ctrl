package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scent of Ivy — Instant {G} (Urza's Destiny, S82 #5621):
//
//	"Reveal any number of green cards in your hand. Target creature gets
//	 +X/+X until end of turn, where X is the number of cards revealed this way."
//
// OldB705ScentOfX: the reveal is chosen as the spell resolves; a hand with no
// green card reveals nothing and the creature gets +0/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6db81e5-b7ca-45dc-a72d-1ff12935a3e4",
		Name:         "Scent of Ivy",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.OldB705ScentOfX("Scent of Ivy", "G", 1, "choose-cards/s82-old-b705-scent-of-ivy"),
	})
}
