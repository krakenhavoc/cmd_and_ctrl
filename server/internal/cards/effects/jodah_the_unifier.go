package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jodah, the Unifier — Legendary Creature — Human Wizard {W}{U}{B}{R}{G},
// 5/5 (Final Fantasy, #4462):
//
//	"Legendary creatures you control get +X/+X, where X is the number of
//	 legendary creatures you control.
//	 Whenever you cast a legendary spell from your hand, exile cards from
//	 the top of your library until you exile a legendary nonland card with
//	 lesser mana value. You may cast that card without paying its mana cost.
//	 Put the rest on the bottom of your library in a random order."
//
// The anthem is a layer 7c modifier read on every pass: X counts the
// legendary creatures you control as the layers have made them, Jodah
// himself included, so a lone Jodah is 6/6.
//
// The trigger fires on a legendary spell cast from your hand (the cast event
// names the zone it left, CR 601.2a). The spell's mana value is read as the
// ability triggers, from its mana cost alone with any {X} at the value
// chosen for it (CR 202.3e, the ruling), and rides the item. The walk is
// cascade's body with a different test (ExileUntilCastFreeKeepingForEffect):
// every card is exiled face up, the hit is the first legendary nonland card
// whose mana value is lower (a card whose cost can't be read is never one),
// and a library with no such card is exiled whole and goes back in a random
// order. The free cast is offered during this resolution (CR 608.2g), so
// the cast spell resolves before the spell that triggered this, even if that
// spell is countered. Declined, the card stays in exile (the ruling) and
// only the rest goes to the bottom. X is 0 for a spell with {X} cast free,
// and any additional cost is paid, as the rulings say.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "490a8045-e160-4ccf-a05e-658ad2b7ac2d",
		Name:         "Jodah, the Unifier",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Legendary creatures you control get +X/+X, where X is the number of legendary creatures you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsLegendary() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for i := range g.Battlefield.Cards {
					b := &g.Battlefield.Cards[i]
					if b.Controller == source.Controller && b.IsCreature() && b.IsLegendary() {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := jodahLegendarySpellFromHand(ev, source, g)
				return ok
			},
			Key: jodahLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, jodahLabel)
				if c, ok := jodahLegendarySpellFromHand(ev, source, g); ok {
					if mv, readable := g.ManaValueForEffect(c); readable {
						item.Params.Amount = mv
					}
				}
				return item
			},
			Effect: jodahExile,
		}},
	})
}

const jodahLabel = "Jodah, the Unifier — exile until a legendary nonland card with lesser mana value; you may cast it free"

// jodahLegendarySpellFromHand is "you cast a legendary spell from your
// hand": the spell, and whether it qualifies.
func jodahLegendarySpellFromHand(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller || ev.OldZone != game.ZoneHand {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !isLegendary(&c) {
		return game.Card{}, false
	}
	return c, true
}

// jodahExile walks the library for the hit.
func jodahExile(g *game.Game, item *game.StackItem) error {
	limit := item.Params.Amount
	return g.ExileUntilCastFreeKeepingForEffect(item.Controller, item.SourceCardID, func(c game.Card) bool {
		if c.IsLand() || c.IsToken() || !Legendary()(g, uuid.Nil, c) {
			return false
		}
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv < limit
	})
}
