package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza's Blueprints — Artifact, {6}:
//
//	"Echo {6} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {T}: Draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad75efdb-3e36-4253-be47-1df5a3cf8204",
		Name:         "Urza's Blueprints",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Draw a card.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{Echo("Urza's Blueprints", "{6}")},
	})
}
