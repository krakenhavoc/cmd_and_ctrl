package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kami of Transience — Creature — Spirit {1}{G}, 2/2 (Kamigawa: Neon
// Dynasty, #5493):
//
//	"Trample
//	 Whenever you cast an enchantment spell, put a +1/+1 counter on this
//	 creature.
//	 At the beginning of each end step, if an enchantment was put into your
//	 graveyard from the battlefield this turn, you may return this card
//	 from your graveyard to your hand."
//
// The cast trigger is Electrostatic Infantry's. The end-step trigger works
// from the graveyard (Dungeon Crawler's InGraveyard + Optional) on every
// player's end step. Its intervening "if" (CR 603.4) reads the turn tally
// of enchantments put into the card's owner's graveyard from the
// battlefield (PlayerTurnTally.EnchantmentsToGraveyard, #5493), at trigger
// time and again on resolution; the Kami itself counts if it is an
// enchantment, which it is not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7baffe00-78a4-4675-9c74-29eb3882bd01",
		Name:            "Kami of Transience",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Enchantment(),
				"Kami of Transience — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			InGraveyard(Optional(On(game.EventBeginEndStep,
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return g.EnchantmentToGraveyardThisTurn(source.Owner)
				},
				"Kami of Transience — return this card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if !g.EnchantmentToGraveyardThisTurn(item.Controller) { // CR 603.4
						return nil
					}
					return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
				}), "Return Kami of Transience from your graveyard to your hand?")),
		},
	})
}
