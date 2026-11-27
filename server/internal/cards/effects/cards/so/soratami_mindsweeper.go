package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soratami Mindsweeper — Creature — Moonfolk Wizard {3}{U}, 1/4 (Betrayers
// of Kamigawa, #5602):
//
//	"Flying
//	 {2}, Return a land you control to its owner's hand: Target player
//	 mills two cards."
//
// Oboro Breezecaller's cost (a land you control, returned as the ability is
// activated) with Millstone's effect: the player is picked at announce and
// re-checked at resolution, and a library with fewer than two cards mills
// what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2ef6f4d6-f755-4402-8339-2a9b6c86f96c",
		Name:            "Soratami Mindsweeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Return a land you control to its owner's hand: Target player mills two cards.",
			Cost: effects.Plus(
				effects.ManaCost("{2}"),
				effects.ReturnAPermanentToHand("a land you control", effects.Land()),
			),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.B31MillChosenPlayer(2),
		}},
	})
}
