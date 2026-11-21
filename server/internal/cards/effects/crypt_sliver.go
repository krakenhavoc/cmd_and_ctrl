package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const cryptSliverGrant = "crypt-sliver/regenerate-target-sliver"

// Crypt Sliver — Creature — Sliver {1}{B}, 1/1 (Commander Masters):
//
//	"All Slivers have "{T}: Regenerate target Sliver.""
//
// Clot Sliver's shape (a bundle of one activated ability granted to every
// Sliver, any player's), with Boneknitter's target clause: the ability
// targets any Sliver on the battlefield, not only its own source. The {T} is
// a creature's, so summoning sickness applies to each Sliver that has it
// (CR 302.6), which the engine reads off the type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0bc2d60-20d0-4ca2-be86-65b297226ac7",
		Name:         "Crypt Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: cryptSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}: Regenerate target Sliver.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    TapCost(),
				Targets: TargetPermanent("target Sliver", HasSubtype("Sliver")),
				Effect:  regenerateTheTargetPermanent,
			}},
			Text: "{T}: Regenerate target Sliver.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, cryptSliverGrant),
		},
	})
}
