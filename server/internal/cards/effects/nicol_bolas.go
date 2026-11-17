package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nicol Bolas — Legendary Creature — Elder Dragon
// {2}{U}{U}{B}{B}{R}{R}, 7/7 (Legends, #5639):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice Nicol Bolas unless you
//	 pay {U}{B}{R}.
//	 Whenever Nicol Bolas deals damage to an opponent, that player
//	 discards their hand."
//
// The upkeep payment is UpkeepSacrificeUnlessYouPay. The discard trigger
// is Looter il-Kor's (WheneverThisDealsDamageToAnOpponent): any damage,
// combat or not, to a player other than its controller. "That player" is
// the damage event's target, and the whole hand goes, as Myojin of
// Night's Reach discards it (discardWholeHand).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e26bf6a9-b31c-4bc0-b55b-c01f2f69be6b",
		Name:            "Nicol Bolas",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Nicol Bolas", "{U}{B}{R}"),
			WheneverThisDealsDamageToAnOpponent("Nicol Bolas — that player discards their hand",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					_, err := discardWholeHand(g, item.Trigger.Event.Target)
					return err
				}),
		},
	})
}
