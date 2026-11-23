package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fearless Pup — Creature — Wolf {R}, 1/1:
//
//	"First strike
//	 Boast — {2}{R}: This creature gets +2/+0 until end of turn. (Activate only if
//	 this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61295adf-9a58-479a-b6e5-403aa0876c22",
		Name:            "Fearless Pup",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{
			effects.BoastAnswering(game.AnswerPump, "{2}{R}: This creature gets +2/+0 until end of turn.",
				effects.ManaCost("{2}{R}"), effects.ThisGetsUntilEndOfTurn(2, 0, "Fearless Pup — +2/+0")),
		},
	})
}
