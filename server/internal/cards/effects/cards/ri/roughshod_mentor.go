package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roughshod Mentor — Creature — Giant Warrior {5}{G}, 5/4 (Shadowmoor, #5591):
//
//	"Green creatures you control have trample."
//
// A layer 6 keyword grant over the green creatures its controller controls,
// itself included, read from current colours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec43fb64-122d-4fc5-9fe0-b7913373f7a4",
		Name:         "Roughshod Mentor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB540ColorCreaturesYouControlHave("G", "trample")},
	})
}
