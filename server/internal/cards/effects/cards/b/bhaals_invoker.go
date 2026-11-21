package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bhaal's Invoker — Creature — Dragon Shaman {2}{R}, 4/2 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Scorching Ray — {8}: This creature deals 4 damage to each opponent."
//
// One activated ability: the damage is dealt by this creature to every opponent
// of its controller at once (CR 120.3), a source that has left being its last-
// known information (CR 608.2h). "Scorching Ray" is an ability word with no
// rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92512acd-bf1e-4133-93cd-6c6af5affb95",
		Name:         "Bhaal's Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{8}: This creature deals 4 damage to each opponent.",
			Cost:  effects.ManaCost("{8}"),
			// Damage to each opponent answers nothing in response.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 4)
			},
		}},
	})
}
