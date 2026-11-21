package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minotaur Skullcleaver — Creature — Minotaur Berserker {2}{R}, 2/2 (Theros, #5561):
//
//	"Haste
//	 When this creature enters, it gets +2/+0 until end of turn."
//
// The enters trigger pumps the Minotaur itself until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd36aea8-9f1c-4daf-86f6-ab9307279ce7",
		Name:            "Minotaur Skullcleaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Minotaur Skullcleaver — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Minotaur Skullcleaver — +2/+0")),
		},
	})
}
