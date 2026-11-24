package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evin, Waterdeep Opportunist — Legendary Creature — Human Rogue {3}{B}, 2/4:
//
//	"Menace
//	 Ward—Sacrifice a creature.
//	 Evin gets +2/+0 for each Treasure you control.
//	 Whenever one or more players sacrifice one or more creatures, you create a
//	 tapped Treasure token. This ability triggers only once each turn."
//
// The ward is Sauron's sacrifice cost (WardSacrifice), the counter-unless-paid
// of CR 702.21a. The bonus is Grand Nagus Zek's layer 7c read of the board, +2/+0
// for each Treasure artifact its controller controls, so a Treasure cracked for
// mana shrinks him at once. The trigger is Zek's too: any player's sacrifice of
// a creature, a cost or an effect, the tapped Treasure for Evin's controller,
// one trigger for a batch of sacrifices (OncePerBatch) and only once each turn,
// counted when it is put on the stack, so a countered one still used up the turn.
//
// No simplification.
func init() {
	const label = "Evin, Waterdeep Opportunist — create a tapped Treasure token"
	effects.Register(effects.Spec{
		OracleID:        "f688f839-c7c2-45a9-8c60-9fdb030803e6",
		Name:            "Evin, Waterdeep Opportunist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:     "Evin gets +2/+0 for each Treasure you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				treasure := source.TextSubtype("Treasure")
				n := effects.CountPermanentsYouControl(g, source, func(o *game.Card) bool { return o.HasSubtype(treasure) })
				c.Power += 2 * n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardSacrifice("a creature", effects.Creature()), "Evin, Waterdeep Opportunist — ward—sacrifice a creature"),
			effects.OncePerBatch(effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.RoeB1APlayerSacrificedACreature(ev, source, lki, g) && !effects.B11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, func(g *game.Game, item *game.StackItem) error {
				treasure := effects.TreasureToken()
				treasure.Tapped = true
				return effects.CreateToken{Template: treasure, N: 1}.Apply(effects.NewContext(g, item))
			})),
		},
	})
}
