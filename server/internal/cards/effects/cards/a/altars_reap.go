package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Altar's Reap — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards."
//
// Village Rites at one mana more — printed years earlier, and still
// played in lists that want a second copy of the effect. Identical
// machinery; kept as its own card because the extra mana genuinely
// matters to the decks running both.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6a125750-2b8c-4f9d-8173-ac8d14c91ddb",
		Name:           "Altar's Reap",
		Purpose:        game.Purpose{Draws: 2, Answers: game.AnswerSacOutlet}, // #2877: the creature sacrifice is the answer (ADR 0142 §3)
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
