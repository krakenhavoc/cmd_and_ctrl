package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hibernation Sliver — Creature — Sliver {U}{B}, 2/2 (Commander Masters):
//
//	"All Slivers have "Pay 2 life: Return this permanent to its owner's
//	 hand.""
//
// A granted activated ability (ADR 0093) on every Sliver at the table, this
// one included. Its controller pays the life (CR 119.4) and may do it at
// instant speed, so a Sliver returns itself in response to removal. The
// target-free ability returns the permanent that has it, wherever its
// controller is.
//
// No simplification.
const hibernationSliverGrant = "hibernation-sliver/return"

func init() {
	Register(Spec{
		OracleID:     "edb8a6d6-9e25-4048-a842-e854eece42c2",
		Name:         "Hibernation Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: hibernationSliverGrant,
			Activated: []ActivatedAbility{{
				Label:   "Pay 2 life: Return this permanent to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    PayLife(2),
				Effect:  returnThisPermanentToOwnersHand,
			}},
			Text: "Pay 2 life: Return this permanent to its owner's hand.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}}, hibernationSliverGrant),
		},
	})
}
