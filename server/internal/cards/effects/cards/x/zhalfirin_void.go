package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zhalfirin Void — Land:
//
//	"When this land enters, scry 1.
//	 {T}: Add {C}."
//
// The Temple cycle's enters-trigger scry (temples.go) on a colourless
// land that enters untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "13aab4fc-4c89-45e6-8275-b074b00d0ee9",
		Name:          "Zhalfirin Void",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Scry: 1},
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Zhalfirin Void — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
