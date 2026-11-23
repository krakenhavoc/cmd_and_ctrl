package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Punk Frogs — {3}{G/U}{G/U} Creature — Frog Mutant Rebel 4/5 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Ward {3}"
//
// Rimeshield Frost Giant's card in another frame: ward is the triggered
// ability ward.go builds (CR 702.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfb13eca-55fd-4ac6-aee3-2bdbfe9e1c4e",
		Name:         "Punk Frogs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Punk Frogs — ward {3}"),
		},
	})
}
