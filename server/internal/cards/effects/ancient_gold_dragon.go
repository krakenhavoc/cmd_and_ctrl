package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Gold Dragon — "Whenever this creature deals combat damage to a
// player, roll a d20. You create a number of 1/1 blue Faerie Dragon creature
// tokens with flying equal to the result."
func init() {
	Register(Spec{
		OracleID:        "44cb725a-72fc-4e0d-b966-f0b6d33f6b79",
		Name:            "Ancient Gold Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Ancient Gold Dragon — roll a d20 and create Faerie Dragons", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				results, err := rollDice(ctx, 20, 1)
				if err != nil || len(results) == 0 {
					return err
				}
				t := TokenCard("1/1 blue Faerie Dragon with flying")
				return g.CreateTokenForEffect(item.Controller, t, results[0])
			}),
		},
	})
}
