package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wistfulness — Creature — Elemental Incarnation {3}{G/U}{G/U}, 6/5
// (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters, if {G}{G} was spent to cast it, exile
//	 target artifact or enchantment an opponent controls.
//	 When this creature enters, if {U}{U} was spent to cast it, draw two
//	 cards, then discard a card.
//	 Evoke {G/U}{G/U} (You may cast this spell for its evoke cost. If
//	 you do, it's sacrificed when it enters.)"
//
// Deceit's cycle-mate: two enters triggers behind intervening ifs
// (CR 603.4) on the mana spent (deceitSpent). The loot is Three Steps
// Ahead's body (drawThenDiscard): the discard is asked of the hand after
// the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "e5e8fce4-c97a-4ae4-9583-003c565e3bd6",
		Name:           "Wistfulness",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{G/U}{G/U}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, deceitSpent("G")),
					"Wistfulness — exile target artifact or enchantment an opponent controls",
					func(g *game.Game, item *game.StackItem) error {
						return exileTheFirstTargetIfLegal(NewContext(g, item))
					}),
				TargetPermanent("target artifact or enchantment an opponent controls",
					Or(Artifact(), Enchantment()), OpponentControls())),
			On(game.EventETB, AllOf(Self, deceitSpent("U")),
				"Wistfulness — draw two cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return drawThenDiscard(g, item.Controller, item.SourceCardID, 2, 1, "")
				}),
		},
	})
}
