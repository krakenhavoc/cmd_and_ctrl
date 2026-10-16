package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intet, the Dreamer — Legendary Creature — Dragon {3}{G}{U}{R}, 6/6:
//
//	"Flying
//	 Whenever Intet deals combat damage to a player, you may pay {2}{U}.
//	 If you do, exile the top card of your library face down. You may
//	 look at that card for as long as it remains exiled. You may play
//	 that card without paying its mana cost for as long as Intet remains
//	 on the battlefield."
//
// The payment is asked as the trigger resolves (MayPay); declining it
// does nothing. Paid, the top card goes to exile face down where only
// its owner may look at it (FaceDownPermitted, which is the "you may
// look at that card for as long as it remains exiled"), and the free
// play is a permission whose window is "for as long as Intet remains on
// the battlefield" (#3192, ADR 0066's second 2026-10-13 amendment:
// game.ForAsLongAs + WhileSourceOnBattlefield, read by the one liveness
// test). The window is measured when the card is exiled, so an Intet
// that left in response leaves the card exiled and unplayable (CR
// 611.2b), and one that leaves later ends it for good, even if Intet
// returns (a new object, CR 400.7). A land is played (it spends a land
// drop); a spell costs {0} in place of its mana cost, with any
// additional costs still owed (CR 118.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9518ffbc-bb3d-4f48-a900-f5bad8302a87",
		Name:            "Intet, the Dreamer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Intet, the Dreamer — you may pay {2}{U} to exile the top card of your library face down", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}{U}",
					Question: "Intet, the Dreamer — pay {2}{U} to exile the top card of your library face down?",
					OnPay: func(ctx *Context) error {
						window := forAsLongAsThisRemainsWindow(ctx.Game, ctx.Item)
						return ExileTopWithPermission{
							From: ctx.Item.Controller, GrantTo: ctx.Item.Controller, N: 1,
							FaceDown: true,
							Free:     true,
							Window:   &window,
						}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
