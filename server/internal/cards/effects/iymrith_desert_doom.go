package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iymrith, Desert Doom — Legendary Creature — Dragon {3}{U}{U}, 5/5
// (Adventures in the Forgotten Realms):
//
//	"Flying
//	 Iymrith has ward {4} as long as it's untapped.
//	 Whenever Iymrith deals combat damage to a player, draw a card. Then if
//	 you have fewer than three cards in hand, draw cards equal to the
//	 difference."
//
// The ward is the catalog's granted ward with the narrowest predicate,
// Iymrith itself while it is untapped, read as a targeting spell or ability
// would trigger it (CR 702.21a), so tapping it (attacking) drops the ward at
// once. The draw is two steps in one resolution: one card, then the hand is
// counted AFTER it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d5efba91-8400-41a3-91c6-cb5e81f07976",
		Name:            "Iymrith, Desert Doom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{4}"), "Iymrith, Desert Doom — ward {4} while untapped",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && !target.Tapped
				}),
			WheneverThisDealsCombatDamageToAPlayer("Iymrith, Desert Doom — draw a card, then up to three in hand",
				Do(Instruction(func(ctx *Context) error {
					if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					me := ctx.PlayerByID(ctx.Controller())
					if me == nil {
						return nil
					}
					if diff := 3 - len(me.Hand.Cards); diff > 0 {
						return DrawCards{Player: ctx.Controller(), N: diff}.Apply(ctx)
					}
					return nil
				}))),
		},
	})
}
