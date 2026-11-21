package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wipe Clean — Instant {1}{W} (Scourge):
//
//	"Exile target enchantment.
//	 Cycling {3} ({3}, Discard this card: Draw a card.)"
//
// The shared exile of the first legal card target, over enchantments, and
// cycling as the keyword's own hand ability (CR 702.29a). An enchantment
// that left in response leaves nothing to exile (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5cbc4f3-4e73-405f-9daf-e5b7c73fa6ee",
		Name:         "Wipe Clean",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{3}")},
	})
}
