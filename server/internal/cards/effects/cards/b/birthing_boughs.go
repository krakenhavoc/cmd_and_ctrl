package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Birthing Boughs — Artifact {3} (Modern Horizons):
//
//	"{4}, {T}: Create a 2/2 colorless Shapeshifter creature token with
//	 changeling. (It is every creature type.)"
//
// Irregular Cohort's token behind a mana-and-tap cost: the Shapeshifter is a
// 2/2 colorless token that carries the changeling keyword, so it is every
// creature type (CR 702.73a). The artifact itself taps, so it has no summoning
// sickness to wait out (CR 302.6 is for creatures).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6836d78-ca71-4413-aced-d4de19976221",
		Name:         "Birthing Boughs",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Create a 2/2 colorless Shapeshifter creature token with changeling.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.ColorlessShapeshifterToken(), N: 1}),
		}},
	})
}
