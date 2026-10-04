package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master of the Feast — Enchantment Creature — Demon {1}{B}{B}, 5/5:
//
//	"Flying
//	 At the beginning of your upkeep, each opponent draws a card."
//
// Mandatory, and every opponent draws, in seat order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01dc86f3-5ebb-4b10-bf68-8fc9f232c724",
		Name:            "Master of the Feast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Master of the Feast — each opponent draws a card", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var err error
				eachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
					err = DrawCards{Player: opp, N: 1}.Apply(ctx)
					return err != nil
				})
				return err
			}),
		},
	})
}
