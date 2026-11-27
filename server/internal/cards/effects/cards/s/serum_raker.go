package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serum Raker — Creature — Phyrexian Drake {2}{U}{U}, 3/2 (Mirrodin Besieged,
// S82 OLD b467, #5576):
//
//	"Flying
//	 When this creature dies, each player discards a card."
//
// Each player in turn order from the active player (CR 101.4) chooses and
// discards one card; a player with an empty hand discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7bb7b937-a7ad-4f5c-b173-df71556c48e9",
		Name:            "Serum Raker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Serum Raker — each player discards a card", func(g *game.Game, item *game.StackItem) error {
				g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
					Source: item.SourceCardID, N: 1, Question: "Serum Raker — discard a card",
				})
				return nil
			}),
		},
	})
}
