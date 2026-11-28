package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Whale — Creature — Whale {5}{U}{U}, 5/5 (Urza's Saga, S82 OLD batch
// 730, #5624):
//
//	"When this creature enters, untap up to seven lands."
//
// Peregrine Drake's trigger with seven: the controller untaps up to seven
// tapped lands of their own, and is asked only when a land is tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c43ce30a-0346-479d-8dbc-56c3872ff82a",
		Name:         "Great Whale",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Great Whale — untap up to seven lands", effects.Do(effects.UntapUpToLands{
				N:        7,
				Question: "Great Whale — untap up to seven lands",
			})),
		},
	})
}
