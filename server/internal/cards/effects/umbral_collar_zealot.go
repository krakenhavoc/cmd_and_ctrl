package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "12db6263-75c2-442f-a1a5-7af7915f8f9f",
		Name:         "Umbral Collar Zealot",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature or artifact: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
