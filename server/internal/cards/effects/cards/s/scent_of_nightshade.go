package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scent of Nightshade — Instant {1}{B} (Urza's Destiny, S82 #5621):
//
//	"Reveal any number of black cards in your hand. Target creature gets
//	 -X/-X until end of turn, where X is the number of cards revealed this way."
//
// OldB705ScentOfX: the reveal is chosen as the spell resolves; a hand with no
// black card reveals nothing and the creature gets -0/-0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e88f39d-6177-42d9-b756-144219dc0961",
		Name:         "Scent of Nightshade",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetHarm}),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.OldB705ScentOfX("Scent of Nightshade", "B", -1, "choose-cards/s82-old-b705-scent-of-nightshade"),
	})
}
