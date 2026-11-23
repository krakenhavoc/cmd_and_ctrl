package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collective Blessing — Enchantment {3}{G}{G}{W} (Starter Commander Decks):
//
//	"Creatures you control get +3/+3."
//
// A layer 7c anthem over every creature its controller controls, itself
// not being one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f7049e8-49ed-46da-89f7-1e40aefb3b0c",
		Name:         "Collective Blessing",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.B16Anthem(effects.B16CreaturesYouControl, 3, 3)},
	})
}
