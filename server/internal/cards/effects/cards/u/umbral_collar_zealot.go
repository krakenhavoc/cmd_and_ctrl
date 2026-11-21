package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Umbral Collar Zealot — Creature — Human Cleric {1}{B}, 3/2 (EDHREC
// rank 4488):
//
//	"Sacrifice another creature or artifact: Surveil 1."
//
// An ordinary CR 602 sacrifice-outlet ability with no mana or tap
// component, so it can fire any number of times a turn, at instant
// speed, for as long as there is something else to feed it.
//
// "Another" is object identity (effects.Another, CR 109.1): the Zealot cannot
// sacrifice itself, but can sacrifice a second Umbral Collar Zealot.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12db6263-75c2-442f-a1a5-7af7915f8f9f",
		Name:         "Umbral Collar Zealot",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature or artifact: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact",
					effects.Or(effects.Creature(), effects.Artifact()))),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Surveil{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
