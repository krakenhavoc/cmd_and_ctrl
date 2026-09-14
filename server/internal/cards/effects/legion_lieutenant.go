package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Legion Lieutenant — Creature — Vampire Knight {W}{B}, 2/2 (EDHREC
// rank 3049):
//
//	"Other Vampires you control get +1/+1."
//
// The Vampire lord. TribalAnthem with the "other" and the "you
// control" both printed; effective subtypes, so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5fa1b2f0-3ba4-49db-9cb5-b6130e4c255e",
		Name:         "Legion Lieutenant",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
