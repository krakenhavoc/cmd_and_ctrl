package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b0bc2d60-20d0-4ca2-be86-65b297226ac7",
		Name:         "Crypt Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: cryptSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Regenerate target Sliver.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target Sliver", effects.HasSubtype("Sliver")),
				Effect:  effects.RegenerateTheTargetPermanent,
			}},
			Text: "{T}: Regenerate target Sliver.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, cryptSliverGrant),
		},
	})
}
