package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skull Rend — Sorcery {3}{B}{R} (Return to Ravnica, S82 OLD):
//
//	"Skull Rend deals 2 damage to each opponent. Those players each discard
//	 two cards at random."
//
// One damage event per opponent, then each opponent still in the game
// discards two cards at random (CR 701.9b); a hand of fewer cards is emptied.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d29d59f1-e00c-4217-8185-2085a4142e3b",
		Name:         "Skull Rend",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DamageEachOpponent{N: 2}).Apply(ctx); err != nil {
				return err
			}
			for _, opp := range ctx.Opponents() {
				if err := (effects.DiscardCards{Player: opp, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
