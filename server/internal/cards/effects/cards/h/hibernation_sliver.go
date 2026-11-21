package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "edb8a6d6-9e25-4048-a842-e854eece42c2",
		Name:         "Hibernation Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: hibernationSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "Pay 2 life: Return this permanent to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.PayLife(2),
				Effect:  effects.ReturnThisPermanentToOwnersHand,
			}},
			Text: "Pay 2 life: Return this permanent to its owner's hand.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, hibernationSliverGrant),
		},
	})
}
