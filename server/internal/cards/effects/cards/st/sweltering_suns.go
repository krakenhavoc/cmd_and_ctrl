package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sweltering Suns — Sorcery {1}{R}{R} (Starter Commander Decks):
//
//	"Sweltering Suns deals 3 damage to each creature.
//	 Cycling {3} ({3}, Discard this card: Draw a card.)"
//
// Slagstorm's first mode: one batch of damage to every creature, read as it
// resolves, with the shared cycling ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "526cd570-046e-402a-986c-f6ece68eec47",
		Name:         "Sweltering Suns",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		Activated:    []effects.ActivatedAbility{effects.Cycling("{3}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), 3)
		},
	})
}
