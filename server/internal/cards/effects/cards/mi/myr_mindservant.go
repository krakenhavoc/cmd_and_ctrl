package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myr Mindservant — Artifact Creature — Myr {1}, 1/1 (Mirrodin, S82 OLD,
// #5606):
//
//	"{2}, {T}: Shuffle your library."
//
// The {T} makes it a summoning-sick creature's ability (CR 302.6): it can't
// shuffle the turn it arrives. Shuffling is a plain action on the
// controller's own library, so the ability answers nothing in response to a
// spell (Purpose answers "value").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0ea83ab-f62e-43ee-b7ad-ddbbc8916632",
		Name:         "Myr Mindservant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Shuffle your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ShuffleLibrary{Player: item.Controller}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
