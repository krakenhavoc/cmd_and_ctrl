package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dfb13eca-55fd-4ac6-aee3-2bdbfe9e1c4e",
		Name:         "Punk Frogs",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Punk Frogs — ward {3}"),
		},
	})
}
