package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "13aab4fc-4c89-45e6-8275-b074b00d0ee9",
		Name:          "Zhalfirin Void",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Scry: 1},
		ManaAbilities: []ManaAbility{colorlessTapRow()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Zhalfirin Void — scry 1", Do(Scry{N: 1})),
		},
	})
}
