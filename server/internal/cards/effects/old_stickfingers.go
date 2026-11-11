package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Old Stickfingers — Legendary Creature — Horror {X}{B}{G}, */* (Duskmourn
// Commander):
//
//	"When you cast this spell, reveal cards from the top of your library
//	 until you reveal X creature cards. Put all creature cards revealed this
//	 way into your graveyard, then put the rest on the bottom of your library
//	 in a random order.
//	 Old Stickfingers's power and toughness are each equal to the number of
//	 creature cards in your graveyard."
//
// The cast trigger reads X off the spell as it goes on the stack (Hydroid
// Krasis' shape: the spell's announced X rides the trigger as data, CR 601.2b),
// so the reveal still happens when Old Stickfingers is countered in response.
// The reveal stops at the Xth creature card or when the library runs out; a
// token is revealed but is never a creature card (CR 108.2). The creature
// cards are put into your graveyard (a put, not a mill) and the rest go to the
// bottom in a random order from the game's keyed stream. X is 0 on the
// battlefield and the trigger with X = 0 reveals nothing. The size is a
// characteristic-defining ability (CR 604.3, 613.4a) over the creature cards
// in your graveyard, which is counted as it is now, so the Stickfingers'
// own cards milled by its trigger count as soon as they arrive.
//
// No simplification.
func init() {
	const label = "Old Stickfingers — reveal until X creature cards; put them into your graveyard and the rest on the bottom in a random order"
	Register(Spec{
		OracleID:     "1d85eec2-80db-466c-bd48-d2a12866e8bd",
		Name:         "Old Stickfingers",
		XMatters:     true,
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "Power and toughness are each equal to the number of creature cards in your graveyard.",
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n, _ := cmdr2S30CreatureCardsIn(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: Self,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Controller, item.Owner = ev.Actor, ev.Actor
				if spell := g.StackItemForEffect(source.InstanceID); spell != nil && spell.XValue > 0 {
					item.Params.Amount = spell.XValue
				}
				return item
			},
			Effect: cmdr2S30OldStickfingersReveal,
		}},
	})
}
