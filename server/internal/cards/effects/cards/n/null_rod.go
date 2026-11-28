package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Null Rod — Artifact {2} (Weatherlight, S82 OLD b776):
//
//	"Activated abilities of artifacts can't be activated."
//
// Stony Silence in artifact form, mana abilities included (CR 605.1a: a mana
// ability is an activated ability); Null Rod's own abilities are none. Loyalty,
// triggered and static abilities are untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f83ca86-e23d-40f7-8085-6928d8cfef9b",
		Name:         "Null Rod",
		Completeness: effects.CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			effects.SourcesCantActivate(
				"Null Rod — activated abilities of artifacts can't be activated.",
				effects.Artifact(), false),
		},
	})
}
