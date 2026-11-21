package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinuous Vermin — {1}{B} Creature — Rat Horror 2/2 (Jumpstart 2022, #5948):
//
//	"{3}{B}{B}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)
//	 As long as this creature is monstrous, it has menace."
//
// Fleecemane Lion's monstrosity and monstrous keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "658f5cfb-326d-4da9-a39b-2ecd393a1f54",
		Name:         "Sinuous Vermin",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{3}{B}{B}"), 3)},
		Static:       []game.StaticAbility{effects.MonstrousKeywords("menace")},
	})
}
