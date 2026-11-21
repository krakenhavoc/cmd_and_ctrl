package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const radagastLabel = "Radagast the Brown — look at the top X cards of your library, where X is that creature's mana value; you may reveal a creature card that doesn't share a creature type with a creature you control and put it into your hand"

// Radagast the Brown — Legendary Creature — Avatar Wizard {2}{G}{G}, 2/5
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever Radagast or another nontoken creature you control enters,
//	 look at the top X cards of your library, where X is that creature's
//	 mana value. You may reveal a creature card that doesn't share a
//	 creature type with a creature you control from among those cards and
//	 put it into your hand. Put the rest on the bottom of your library in a
//	 random order."
//
// Brightwood Tracker's look-and-take (LookAtTopThenMayTakeToHand) on
// Rhythm of the Wild's "this or another nontoken creature" watch. X is
// the entering creature's mana value, read as the trigger is put on the
// stack (a permanent's X is 0, CR 107.3f) and carried on the item; the
// "doesn't share a creature type" test is made as the card is chosen, over
// the creatures its controller controls then, Radagast and the creature
// that entered included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfa47d0d-17f0-4a12-89d3-0d3c7f3db8c6",
		Name:         "Radagast the Brown",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			Key:       radagastLabel,
			AppliesTo: effects.Cmdr2S33ThisOrAnotherNontokenCreatureYouControlEntered,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, radagastLabel)
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					item.Params.Amount = c.ManaValue()
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return effects.LookAtTopThenMayTakeToHand(item.Params.Amount, effects.GapB125NoSharedCreatureType(), 1,
					"Radagast the Brown — you may reveal a creature card that doesn't share a creature type with a creature you control and put it into your hand")(g, item)
			},
		}},
	})
}
