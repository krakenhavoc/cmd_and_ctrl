package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Floodbringer — Creature — Moonfolk Wizard {1}{U}, 1/2 (Betrayers of
// Kamigawa, #5602):
//
//	"Flying
//	 {2}, Return a land you control to its owner's hand: Tap target land."
//
// Oboro Breezecaller's cost with a tap instead of an untap. A land returned
// for the cost is gone before the target is checked, so it cannot be the
// land tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6c8bba8b-44de-4bbf-bf6a-3045e7313995",
		Name:            "Floodbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Return a land you control to its owner's hand: Tap target land.",
			Cost: effects.Plus(
				effects.ManaCost("{2}"),
				effects.ReturnAPermanentToHand("a land you control", effects.Land()),
			),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.TapEachLegalTarget,
		}},
	})
}
