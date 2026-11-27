package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire Ants — Creature — Insect 2/1 {2}{R} (Urza's Saga, S82 OLD b723):
//
//	{T}: This creature deals 1 damage to each other creature without flying.
//
// A tap ability that damages every other creature without flying, friend or foe, as one instance; this creature is spared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c23543b7-f637-4a7c-9b72-ec411cd5488e",
		Name:         "Fire Ants",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each other creature without flying.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying"), effects.OtherThan(ctx.Source())), 1)
			},
		}},
	})
}
