package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siegebreaker Giant — Creature — Giant Warrior {3}{R}{R}, 6/3
// (Core Set 2019, #5527):
//
//	"Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	{3}{R}: Target creature can't block this turn."
//
// Trample, and Chainwhip Cyclops's ability: one target creature can't
// block this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcd281e3-f336-449f-a887-4b98dc14dfe7",
		Name:            "Siegebreaker Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: Target creature can't block this turn.",
			Cost:    effects.ManaCost("{3}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Siegebreaker Giant — target creature can't block this turn"),
		}},
	})
}
