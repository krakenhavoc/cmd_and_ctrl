package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miriam, Herd Whisperer — Legendary Creature — Human Druid {G}{W}, 3/2
// (Outlaws of Thunder Junction, #4468):
//
//	"During your turn, Mounts and Vehicles you control have hexproof.
//	 Whenever a Mount or Vehicle you control attacks, put a +1/+1 counter
//	 on it."
//
// The grant is Kolodin, Triumph Caster's Mount-or-Vehicle KeywordGrant,
// read only while you are the active player, so on an opponent's turn
// your Mounts and Vehicles can be targeted. A Vehicle that is not a
// creature still has hexproof. The counter goes on the attacker that
// triggered it, recorded as the object it was (Kolodin's stillThatObject
// check): one that left combat but is still on the battlefield gets it,
// one that left the battlefield does not. Miriam is neither a Mount nor
// a Vehicle, so she never triggers her own ability.
//
// No simplification.
func init() {
	const label = "Miriam, Herd Whisperer — put a +1/+1 counter on the attacking Mount or Vehicle"
	mountOrVehicle := func(target *game.Card, source *game.Card) bool {
		return target.HasSubtype(source.TextSubtype("Mount")) || target.HasSubtype("Vehicle")
	}
	Register(Spec{
		OracleID:     "1d88014b-8189-4a88-bee3-c2310d36f6e8",
		Name:         "Miriam, Herd Whisperer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller && mountOrVehicle(target, source) && IsYourTurn(g, source.Controller)
		}, "hexproof")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && mountOrVehicle(&c, source)
			},
			Build: eventCardTriggerItem(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				o := item.Params.Object
				c, ok := g.LookupCardForEffect(o.ID)
				if !ok || c.ObjectEpoch != o.Epoch || !onBattlefield(g, o.ID) {
					return nil
				}
				return AddCounter{Target: o.ID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
