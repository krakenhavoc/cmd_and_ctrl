package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thirsting Bloodlord — Creature — Vampire {2}{B}{B}, 3/3 (Foundations Jumpstart):
//
//	"Other Vampires you control get +1/+1."
//
// Lord of the Accursed's anthem over Vampires you control, itself excluded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b04ba676-a1f5-4216-9d93-3c475b504bfa",
		Name:         "Thirsting Bloodlord",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
