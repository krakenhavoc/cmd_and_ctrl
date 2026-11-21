package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magus of the Wheel — Creature — Human Wizard {2}{R}, 3/3 (EDHREC
// rank 1148):
//
//	"{1}{R}, {T}, Sacrifice this creature: Each player discards their
//	 hand, then draws seven cards."
//
// Wheel of Fortune on legs. A three-component activated ability —
// mana, tap (so summoning sickness applies, CR 302.6) and sacrifice
// the source — paid at announce, so the Magus is in the graveyard
// before the wheel resolves and a dies-payoff triggers above it. The
// body is Wheel of Fortune's: every discard before any draw, each
// discard its own event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df13dafb-f82e-42ac-b6dc-efe82af5db58",
		Name:         "Magus of the Wheel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}, Sacrifice this creature: Each player discards their hand, then draws seven cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.B10EachPlayerWheels,
		}},
	})
}
