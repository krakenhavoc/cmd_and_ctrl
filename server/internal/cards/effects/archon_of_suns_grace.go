package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of Sun's Grace — Creature — Archon {2}{W}{W}, 3/4 (EDHREC
// rank 1195):
//
//	"Flying
//	 Lifelink
//	 Pegasus creatures you control have lifelink.
//	 Constellation — Whenever an enchantment you control enters,
//	 create a 2/2 white Pegasus creature token with flying."
//
// The enchantress deck's flying army. Flying and lifelink ride
// PrintedKeywords; the Pegasus grant is a Layer 6 static (Lord of
// Atlantis' shape, without the "other" — the Archon is not a Pegasus,
// so nothing changes for it); constellation is the landfall trigger
// shape with "enchantment" for "land", one trigger per enchantment.
// The token's flying is real, carried on its Keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48a99dce-0aa9-4aac-81df-cec5f94c639d",
		Name:            "Archon of Sun's Grace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && target.HasSubtype("Pegasus")
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == "lifelink" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "lifelink")
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsEnchantment()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Archon of Sun's Grace — create a 2/2 Pegasus with flying (constellation)",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: b10WhitePegasusToken(), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
