package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benalish Marshal — Creature — Human Knight {W}{W}{W}, 3/3 (Dominaria, #5530):
//
//	"Other creatures you control get +1/+1."
//
// A layer 7c anthem over the other creatures you control; the Marshal
// itself is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2cc439e8-d112-44e6-bc5a-6e99333c519a",
		Name:         "Benalish Marshal",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.OtherCreaturesYouControl, 1, 1)},
	})
}
