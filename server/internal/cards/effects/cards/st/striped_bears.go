package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Striped Bears — Creature — Bear {3}{G}, 2/2 (Weatherlight, #5629):
//
//	"When this creature enters, draw a card."
//
// Elvish Visionary's enters trigger: it goes on the stack and the
// controller draws when it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac856e91-1b88-464b-9420-dc311ce814ea",
		Name:         "Striped Bears",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Striped Bears — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
