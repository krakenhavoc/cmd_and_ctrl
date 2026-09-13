package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// No Mercy — Enchantment {2}{B}{B} (EDHREC rank 1469):
//
//	"Whenever a creature deals damage to you, destroy it."
//
// The rattlesnake. Any creature's damage to the controller — combat
// or not, an opponent's creature or the controller's own, as printed
// — triggers it; the creature rides the trigger as a copied ID and
// is destroyed at resolution if it is still on the battlefield. Not
// targeted, so hexproof does not save it; indestructible does, as
// printed. One trigger per damaging creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9538d53-480b-481a-abbf-83ab17e1a45b",
		Name:         "No Mercy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b13CreatureDealtDamageToYou(ev, source, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				attacker, _ := b13CreatureDealtDamageToYou(ev, source, g)
				return game.NewTriggeredItem(source, "No Mercy — destroy the creature that damaged you",
					func(g *game.Game, item *game.StackItem) error {
						if z := g.FindCardZoneForEffect(attacker); z == nil || z.Kind != game.ZoneBattlefield {
							return nil
						}
						return DestroyTarget{Target: attacker}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
