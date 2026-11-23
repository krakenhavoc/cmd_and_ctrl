package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cache Raiders — Creature — Merfolk Rogue {3}{U}{U}, 4/4 (Eventide, #5590):
//
//	"At the beginning of your upkeep, return a permanent you control to its
//	 owner's hand."
//
// Trusted Advisor's upkeep return over any permanent: no target, so it is
// picked as the trigger resolves, the Raiders themselves included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57947037-91b2-4058-be22-158d3257317e",
		Name:         "Cache Raiders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Cache Raiders — return a permanent you control", effects.Do(effects.ReturnOneYouControl{
				Question: "Cache Raiders — return a permanent you control to its owner's hand",
			})),
		},
	})
}
