package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thirsting Bloodlord — Creature — Vampire {2}{B}{B}, 3/3 (Foundations Jumpstart):
//
//	"Other Vampires you control get +1/+1."
//
// Lord of the Accursed's anthem over Vampires you control, itself excluded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b04ba676-a1f5-4216-9d93-3c475b504bfa",
		Name:         "Thirsting Bloodlord",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
