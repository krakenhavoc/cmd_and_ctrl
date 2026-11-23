package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Fleet Sprinter — Creature — Human Pirate {1}{U}{R}, 2/2 (Rivals of
// Ixalan, #5531):
//
//	"Haste
//	 This creature can't be blocked."
//
// Haste is printed data (CR 702.10b); "can't be blocked" is Talas
// Warrior's self restriction, judged as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d244f4a-1d62-4697-81c7-fbc4de874add",
		Name:            "Storm Fleet Sprinter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
