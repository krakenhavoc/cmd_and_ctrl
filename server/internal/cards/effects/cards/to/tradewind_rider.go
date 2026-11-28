package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tradewind Rider — {3}{U} Creature — Spirit, 1/4 (Tempest, S82 OLD batch 770,
// #5628):
//
//	"Flying
//	 {T}, Tap two untapped creatures you control: Return target permanent to its
//	 owner's hand."
//
// Flying is a printed keyword. The cost is the Rider's own tap and two OTHER
// untapped creatures you control (Myr Turbine's tap-others beside the tap: the
// Rider is already spent, so it can't be one of the two); a summoning-sick
// creature can pay the tap-others part, which is not a {T} cost (CR 302.6). The
// bounce is of any permanent, to its owner's hand, re-checked as the ability
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	const tapTwo = "two untapped creatures you control"
	effects.Register(effects.Spec{
		OracleID:        "d0314da0-da24-4f6e-b1f6-0d50617bcaaa",
		Name:            "Tradewind Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Tap two untapped creatures you control: Return target permanent to its owner's hand.",
			Cost: effects.Plus(effects.TapCost(), game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent(tapTwo, effects.Creature()),
				Label:  tapTwo,
			}}),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
