package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Ritual — Instant {1}{B}:
//
//	"Add {B}{B}{B}.
//	 Threshold — Add {B}{B}{B}{B}{B} instead if there are seven or more
//	 cards in your graveyard."
//
// Dark Ritual with a threshold branch. Threshold is read as the spell
// resolves (SpellThreshold, the same condition Lightning Surge reads),
// so the card itself — on the stack, not in the graveyard — never
// counts toward its own seven.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b5bf1fa-6502-4790-b66b-f0f8504ebc7c",
		Name:         "Cabal Ritual",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if effects.SpellThreshold()(ctx.Game, item) {
				return effects.AddMana{Produced: "{B}{B}{B}{B}{B}"}.Apply(ctx)
			}
			return effects.AddMana{Produced: "{B}{B}{B}"}.Apply(ctx)
		},
	})
}
