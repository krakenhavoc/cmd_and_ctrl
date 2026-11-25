package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legion Guildmage — Creature — Human Wizard {R}{W}, 2/2 (Guilds of Ravnica,
// #5525):
//
//	"{5}{R}, {T}: This creature deals 3 damage to each opponent.
//	 {2}{W}, {T}: Tap another target creature."
//
// Both abilities have {T} in the cost, so summoning sickness applies. The
// second targets a creature other than this one (the tapped creature's own
// untapped state is not required).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb33abbe-b43c-4a74-b2db-cce949f0edf2",
		Name:         "Legion Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{5}{R}, {T}: This creature deals 3 damage to each opponent.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{5}{R}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachOpponent{N: 3}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}{W}, {T}: Tap another target creature.",
				Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect:  effects.TapChosenPermanent,
			},
		},
	})
}
