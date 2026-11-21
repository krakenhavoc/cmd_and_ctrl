package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Blueprints — Artifact, {6}:
//
//	"Echo {6} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {T}: Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad75efdb-3e36-4253-be47-1df5a3cf8204",
		Name:         "Urza's Blueprints",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Urza's Blueprints", "{6}")},
	})
}
