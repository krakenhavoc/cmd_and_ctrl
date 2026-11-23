package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turtle Power! — {2}{G} Enchantment (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flash
//	 Turtles you control get +2/+2."
//
// Flash is the printed keyword. The anthem is a tribal layer 7c pump over
// the Turtles its controller controls (Canopy Tactician's TribalAnthem,
// without "other"), the creature type read through the card's own text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15934a19-a29e-4728-816f-20b610eec377",
		Name:            "Turtle Power!",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Turtle"}, YoursOnly: true}, 2, 2),
		},
	})
}
