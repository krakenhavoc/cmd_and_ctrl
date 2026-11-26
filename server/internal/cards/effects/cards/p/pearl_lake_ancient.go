package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pearl Lake Ancient — Creature — Leviathan {5}{U}{U}, 6/7 (Khans of Tarkir):
//
//	"Flash
//	 This spell can't be countered.
//	 Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1
//	 until end of turn.)
//	 Return three lands you control to their owner's hand: Return this
//	 creature to its owner's hand."
//
// Hullbreaker Horror's flash and uncounterable spell, Ringwarden Owl's prowess
// keyword, and the cost of returning three lands you control (Master
// Transmuter's return-to-hand cost, three at a time), chosen as the cost is
// paid (CR 602.2b). The effect is Savage Knuckleblade's bounce of itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7ae125c-955b-4cfd-9aeb-e6572c0d4d3d",
		Name:            "Pearl Lake Ancient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "prowess"},
		CantBeCountered: true,
		Activated: []effects.ActivatedAbility{{
			Label:   "Return three lands you control to their owner's hand: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ReturnNToHand(3, "three lands you control", effects.Land()),
			Effect:  effects.S82OldB38ReturnThisToOwnersHand,
		}},
	})
}
