package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhoda, Geist Avenger — Legendary Creature — Human Soldier {3}{W}, 3/3
// (Innistrad: Crimson Vow Commander):
//
//	"Partner with Timin, Youthful Geist (When this creature enters, target
//	 player may put Timin into their hand from their library, then
//	 shuffle.)
//	 Vigilance
//	 Whenever a creature an opponent controls becomes tapped, if it isn't
//	 being declared as an attacker, put a +1/+1 counter on Rhoda."
//
// Partner with is the shared entry trigger. The counter trigger watches
// only EventTapCard: the engine taps a declared attacker without that
// event (it emits EventAttack), so an attacker being declared never
// matches, while a creature tapped by an effect or a cost does. A Rhoda
// that has left by resolution gets no counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3579e40c-6ec5-4927-8252-5a188db400ab",
		Name:            "Rhoda, Geist Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Rhoda, Geist Avenger", "Timin, Youthful Geist"),
			On(game.EventTapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller != source.Controller && g.Battlefield.Contains(ev.CardID)
			}, "Rhoda — put a +1/+1 counter on Rhoda",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
