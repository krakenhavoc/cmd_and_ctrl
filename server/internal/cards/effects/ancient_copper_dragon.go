package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Copper Dragon — "Whenever this creature deals combat damage to a
// player, roll a d20. You create a number of Treasure tokens equal to the
// result."
func init() {
	Register(Spec{
		OracleID:        "48daee9d-ddaf-410f-8c3a-12fa1064ab56",
		Name:            "Ancient Copper Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Ancient Copper Dragon — roll a d20 and create Treasures", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				results, err := rollDice(ctx, 20, 1)
				if err != nil || len(results) == 0 {
					return err
				}
				return treasureTokens(g, item.Controller, results[0])
			}),
		},
	})
}
