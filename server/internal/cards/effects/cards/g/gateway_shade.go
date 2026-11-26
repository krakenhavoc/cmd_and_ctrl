package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gateway Shade — Creature — Shade {2}{B}, 1/1 (Gatecrash, S82 OLD #5565):
//
//	"{B}: This creature gets +1/+1 until end of turn.
//	 Tap an untapped Gate you control: This creature gets +2/+2 until end of
//	 turn."
//
// A pump row with no {T} (so it can be activated any number of times), and
// Crackling Perimeter's cost of tapping an untapped Gate: not the {T} symbol
// (CR 302.6), so a Gate that arrived this turn may pay it, and it does not
// target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff0647bc-87d1-45f0-ad5c-63c29174f404",
		Name:         "Gateway Shade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.ThisGetsRow("{B}", 1, 1, "Gateway Shade"),
			{
				Label:   "Tap an untapped Gate you control: This creature gets +2/+2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
				Cost:    effects.TapAnUntapped("an untapped Gate you control", effects.Subtype("Gate")),
				Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Gateway Shade — +2/+2 until end of turn"),
			},
		},
	})
}
