package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anax, Hardened in the Forge — Legendary Enchantment Creature — Demigod
// {1}{R}{R}, */3 (Commander Masters):
//
//	"Anax's power is equal to your devotion to red. (Each {R} in the mana
//	 costs of permanents you control counts toward your devotion to red.)
//	 Whenever Anax or another nontoken creature you control dies, create a
//	 1/1 red Satyr creature token with "This token can't block." If the
//	 creature had power 4 or greater, create two of those tokens instead."
//
// The power is Daxos's characteristic-defining ability over devotion
// (layer 7a, CR 604.3), the colour read through Anax's text. The trigger
// reads the creature as it died (CR 603.10a): it was yours, it was nontoken
// (Anax himself always counts, as printed, token copy or not), and its power
// is the last-known power with every counter, read as the trigger goes on the
// stack and carried on the item (CR 608.2h), so "power 4 or greater" is what
// it was on the battlefield. The Satyr is its own catalog token with the
// printed static.
//
// No simplification.
func init() {
	const label = "Anax, Hardened in the Forge — create a 1/1 red Satyr token that can't block (two if the creature had power 4 or greater)"
	t := effects.On(game.EventLTB, gapB104AnaxOrAnotherNontokenCreatureYouControlDied, label,
		func(g *game.Game, item *game.StackItem) error {
			n := 1
			if item.Params.Amount >= 4 {
				n = 2
			}
			return effects.CreateToken{Template: effects.GapB104SatyrCantBlockToken(), N: n}.Apply(effects.NewContext(g, item))
		})
	t.Build = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = effects.B13LastKnownPower(g, ev.CardID, lki)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:     "e23cd06d-1360-4e5d-aecc-adf902bede69",
		Name:         "Anax, Hardened in the Forge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{{
			Label:    "Anax's power is equal to your devotion to red.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.DevotionTo(g, source.Controller, source.TextColor("R"))
			},
		}},
		Triggered: []game.TriggeredAbility{t},
	})
}

// gapB104AnaxOrAnotherNontokenCreatureYouControlDied is "whenever Anax or
// another nontoken creature you control dies".
func gapB104AnaxOrAnotherNontokenCreatureYouControlDied(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if !effects.ACreatureYouControlDied(ev, source, ch, g) {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	dead, ok := effects.DiedCreature(ev, g)
	return ok && !effects.IsToken(dead)
}
