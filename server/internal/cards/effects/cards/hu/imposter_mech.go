package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const imposterMechCrewGrant = "imposter-mech/crew-three"

// Imposter Mech — Artifact — Vehicle {1}{U}, 3/1 (Neon Dynasty):
//
//	"You may have this Vehicle enter as a copy of a creature an opponent
//	 controls, except it's a Vehicle artifact with crew 3 and it loses all
//	 other card types.
//	 Crew 3"
//
// Machine God's Effigy's copy shape: the copy's card types are REPLACED with
// Artifact (SetCardTypes, CR 707.9b; the creature subtypes go with them),
// then the Vehicle subtype is added, and "crew 3" is a registered grant
// (CR 707.9a) because a copy takes the copied card's abilities and not the
// Mech's printed Crew. Declining the copy leaves the printed 3/1 Vehicle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f434e266-d9d6-4058-96ff-a1bd4029d8dc",
		Name:         "Imposter Mech",
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{{Key: imposterMechCrewGrant, Activated: []effects.ActivatedAbility{imposterMechCrew()}, Text: "Crew 3"}},
		Activated:    []effects.ActivatedAbility{imposterMechCrew()},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Imposter Mech",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return effects.CopyCandidates(g, self, func(c game.Card) bool {
						return c.IsCreature() && c.Controller != controller
					})
				},
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.SetCardTypes("Artifact")
					v.AddSubtype("Vehicle")
					v.GrantAbility(imposterMechCrewGrant)
				},
			),
		},
	})
}

func imposterMechCrew() effects.ActivatedAbility {
	return effects.ActivatedAbility{
		Label:   "Crew 3",
		Cost:    effects.CrewCost(3),
		Purpose: game.Purpose{Answers: game.AnswerAnimate},
		Effect:  effects.CrewEffect("Imposter Mech"),
	}
}
