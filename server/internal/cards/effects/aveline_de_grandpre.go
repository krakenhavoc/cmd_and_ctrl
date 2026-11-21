package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aveline de Grandpré — Legendary Creature — Human Assassin {2}{G}{G}, 3/3
// (Assassin's Creed, #5948):
//
//	"Deathtouch
//	 Whenever a creature you control with deathtouch deals combat damage to
//	 a player, put that many +1/+1 counters on that creature.
//	 Disguise {B}{G} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)"
//
// One trigger per creature that connects. The dealer must have deathtouch
// as it dealt the damage (CR 603.10a's look back is not needed: the
// keyword is read off the dealer when the event is judged, which is as it
// dealt it). The amount is the damage dealt, carried on the stack item
// with the dealer's identity, so the counters go on that same object and
// nothing happens if it has left the battlefield (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Aveline de Grandpré — put that many +1/+1 counters on that creature"
	Register(Spec{
		OracleID:         "5c8c4d53-f364-40da-989b-cfe1e3f79b7d",
		Name:             "Aveline de Grandpré",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{B}{G}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				src, ok := g.LookupCardForEffect(ev.Source)
				return ok && game.HasKeyword(&src, "deathtouch")
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = ev.Amount
				item.Params.Object = game.ObjectRef{ID: ev.Source, Epoch: g.ObjectEpochForEffect(ev.Source)}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ref := item.Params.Object
				c, ok := g.LookupCardForEffect(ref.ID)
				if !ok || c.ObjectEpoch != ref.Epoch || !onBattlefield(g, ref.ID) || item.Params.Amount <= 0 {
					return nil
				}
				return AddCounter{Target: ref.ID, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
