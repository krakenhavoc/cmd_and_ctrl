package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrupted Conviction — Instant {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards."
//
// Village Rites under a different name and printing — the sacrifice
// is a COST, so the creature dies with Corrupted Conviction still on
// the stack: a Blood Artist trigger drains before the cards are
// drawn, and countering the spell doesn't hand the creature back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "b45e35df-9032-4482-89a6-c7c50c6d0a79",
		Name:           "Corrupted Conviction",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
