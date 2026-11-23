package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tolarian Sentinel — Creature — Human Spellshaper {3}{U}, 1/3 (Jumpstart 2022, #5948):
//
//	"Flying
//	 {U}, {T}, Discard a card: Return target permanent you control to its
//	 owner's hand."
//
// Waterfront Bouncer's discard-cost bounce, aimed at a permanent you
// control; the discard is a cost, so it is paid when the ability is
// activated and is not undone if the target is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bef15918-c0e7-4f6d-a589-f91f3fa11ef1",
		Name:            "Tolarian Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: Return target permanent you control to its owner's hand.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetPermanent("target permanent you control", effects.YouControl()),
			Effect:  effects.BounceChosenTarget,
		}},
	})
}
