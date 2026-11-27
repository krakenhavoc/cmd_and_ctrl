package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rending Vines — Instant — Arcane {1}{G}{G} (Saviors of Kamigawa, #5601):
//
//	"Destroy target artifact or enchantment if its mana value is less than
//	 or equal to the number of cards in your hand.
//	 Draw a card."
//
// The mana value and the hand are read as the spell resolves, after the
// target is confirmed legal (CR 608.2b); the destroy is skipped when the
// target costs more, and the draw happens either way. Drawing is last, so the
// hand it counts is the one the spell left behind: the spell itself is on the
// stack, not in the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64ab6dfe-ed87-47ec-afd5-a7f2e7e6bbd2",
		Name:         "Rending Vines",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve:    effects.OldB585RendingVinesResolve,
	})
}
