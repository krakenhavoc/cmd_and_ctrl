package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Archers — Creature — Human Archer {3}{R}, 3/4 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"Reach
//	 {5}: This creature deals 2 damage to each opponent. Create a 2/2 red
//	 Soldier creature token."
//
// The Soldier is the Eternal set's plain 2/2 red Soldier, not the
// firebending one of the main set.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e735849-1930-4f43-86b4-da660ade15bc",
		Name:            "Fire Nation Archers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}: This creature deals 2 damage to each opponent. Create a 2/2 red Soldier creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{5}"),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 2}, effects.CreateToken{Template: effects.TokenCard("2/2 red Soldier"), N: 1}),
		}},
	})
}
