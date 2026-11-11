package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cazur, Ruthless Stalker — Legendary Creature — Human Warrior {3}{G},
// 3/3 (Outlaws of Thunder Junction Commander, #5368):
//
//	"Partner with Ukkima, Stalking Shadow (When this creature enters,
//	 target player may put Ukkima into their hand from their library,
//	 then shuffle.)
//	 Whenever a creature you control deals combat damage to a player,
//	 put a +1/+1 counter on that creature."
//
// Partner with is CR 702.124j's two abilities, as Toothy, Imaginary
// Friend's is. The second trigger fires once per creature dealing combat
// damage (each is its own event, CR 510.2), and the counter goes on the
// creature that dealt it, wherever it now is: a creature that left the
// battlefield before the trigger resolves gets nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ccfc74f7-bf32-4201-ad8f-f552ed7f96f8",
		Name:         "Cazur, Ruthless Stalker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Cazur, Ruthless Stalker", "Ukkima, Stalking Shadow"),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Cazur, Ruthless Stalker — put a +1/+1 counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id := item.Trigger.Event.Source
					if !onBattlefield(g, id) {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}
