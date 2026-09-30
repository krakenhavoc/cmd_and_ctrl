package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Silver Dragon — Creature — Elder Dragon {6}{U}{U}, 8/8:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, roll a
//	 d20. Draw cards equal to the result. You have no maximum hand size
//	 for the rest of the game."
//
// The blue member of the Ancient Dragon cycle, on the same trigger
// shape as Ancient Gold Dragon and Ancient Copper Dragon: one d20 from
// the keyed random stream (ADR 0054), then the payoff.
//
// "You have no maximum hand size for the rest of the game" is the
// player-level grant Finale of Revelation and Sea Gate Restoration
// use (Game.SetMaxHandSizeForEffect), not Spec.NoMaxHandSize: it
// outlives the Dragon, which is what "for the rest of the game" says.
// It is set after the draw, as printed; both happen during the same
// resolution, so nothing can observe the order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "939556c8-e363-4453-a624-ca52b151467e",
		Name:            "Ancient Silver Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Ancient Silver Dragon — roll a d20 and draw that many cards", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				results, err := rollDice(ctx, 20, 1)
				if err != nil || len(results) == 0 {
					return err
				}
				if err := (DrawCards{Player: item.Controller, N: results[0]}).Apply(ctx); err != nil {
					return err
				}
				return g.SetMaxHandSizeForEffect(item.Controller, game.NoMaxHandSize)
			}),
		},
	})
}
