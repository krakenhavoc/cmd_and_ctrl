package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimpse of Nature — Sorcery {G}:
//
//	"Whenever you cast a creature spell this turn, draw a card."
//
// CR 603.7b's repeating delayed trigger (#2169): it fires on every
// creature spell the caster casts until cleanup, goes on the stack above
// the spell and resolves first, and is not tied to the sorcery, which is
// in the graveyard by then. The Glimpse itself is not a creature spell, so
// it never triggers itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fae52ed-1b84-406e-951d-b0d329dc48a7",
		Name:         "Glimpse of Nature",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DelayedOnEvent{
				Label:      "Glimpse of Nature — draw a card",
				On:         []game.EventKind{game.EventCast},
				Condition:  effects.YouNextCastCondition,
				CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Creature"}}},
				Body:       effects.DrawOneBody,
				Repeats:    true,
			}.Apply(ctx)
		},
	})
}
