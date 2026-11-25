package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seven Dwarves — Creature — Dwarf {1}{R}, 2/2 (Secret Lair Drop):
//
//	"This creature gets +1/+1 for each other creature named Seven Dwarves you
//	 control.
//	 A deck can have up to seven cards named Seven Dwarves."
//
// The bonus is a layer-7c count of the other creatures with the printed name
// its controller has (Relentless Rats' shape, limited to your side). The deck
// rule is CR 113.6n's, read off the oracle text by internal/deck.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "526ca4a9-3f50-4f7a-8169-2bda95792401",
		Name:         "Seven Dwarves",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.S82OldB024PlusOnePerOtherNamedYouControl("Seven Dwarves")},
	})
}
