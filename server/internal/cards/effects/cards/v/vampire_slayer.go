package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Slayer — Creature — Human Soldier {1}{W}, 2/2 (Innistrad:
// Crimson Vow, #5495):
//
//	"Whenever this creature deals damage to a Vampire, destroy that
//	 creature."
//
// Spider-Slayer's trigger (any damage, not only combat damage) with the
// creature type read as a word of this card's text (CR 612.2) and the
// Vampire judged through the layers as the damage was dealt. The trigger
// names the damaged creature as the object it was when the trigger fired,
// so one that died to the damage, or died and came back, is not destroyed
// (CR 400.7).
//
// No simplification.
func init() {
	const label = "Vampire Slayer — destroy the Vampire it dealt damage to"
	effects.Register(effects.Spec{
		OracleID:     "bae41d41-9dbe-4d6b-bf5d-8a7d7041462c",
		Name:         "Vampire Slayer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source != source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				if !ok || !c.IsCreature() || !c.HasSubtypeWord(source.TextSubtype("Vampire")) {
					return false
				}
				z := g.FindCardZoneForEffect(ev.Target)
				return z != nil && z.Kind == game.ZoneBattlefield
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: effects.DestroyDamagedCreature,
		}},
	})
}
