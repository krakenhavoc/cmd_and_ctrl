package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Athreos, Shroud-Veiled — Legendary Enchantment Creature — God {4}{W}{B},
// 4/7 (Commander, #5368):
//
//	"Indestructible
//	 As long as your devotion to white and black is less than seven,
//	 Athreos isn't a creature.
//	 At the beginning of your end step, put a coin counter on another
//	 target creature.
//	 Whenever a creature with a coin counter on it dies or is put into
//	 exile, return that card to the battlefield under your control."
//
// The God clause is Xenagos, God of Revels' (godUnlessDevotionToColors,
// a layer-4 type loss, so it stays an enchantment). The coin counter goes
// on another creature of any controller, chosen as the trigger goes on
// the stack (CR 603.3d). The last ability watches every creature leaving
// the battlefield for a graveyard or exile, and asks whether it had a
// coin counter as it last existed there (CR 603.10a,
// LastKnownCountersForEffect); the card comes back as a new object under
// Athreos' controller's control (CR 400.7), with no counter, so it does
// not return a second time. A token is never a card that can return.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9aacc1ce-8d0d-4a24-9a15-74635f2f127d",
		Name:            "Athreos, Shroud-Veiled",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{godUnlessDevotionToColors(7, "W", "B")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Athreos, Shroud-Veiled — put a coin counter on another target creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return AddCounter{Target: id, Kind: stdB31CoinCounter, N: 1}.Apply(ctx)
					}),
				Another(TargetCreature("another target creature"))),
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || (ev.NewZone != game.ZoneGraveyard && ev.NewZone != game.ZoneExile) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && leftAsType(ev, c, "creature") && g.LastKnownCountersForEffect(ev.CardID)[stdB31CoinCounter] > 0
			}, "Athreos, Shroud-Veiled — return that card to the battlefield under your control",
				cmdr2B09ReturnTheEventCardUnderYourControl),
		},
	})
}
