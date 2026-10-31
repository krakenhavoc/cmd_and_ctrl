package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Lies of Loki — Creature — Illusion Villain {1}{U}, 1/3 (Marvel
// Super Heroes Commander, #4455):
//
//	"This creature gets +1/+0 for each other Illusion you control.
//	 When this creature dies, draw a card."
//
// The static is Pride Sovereign's shape (a layer 7c count of the other
// creatures of one type, read through the source's text), on power only.
// The death trigger is WhenThisDies with a DrawCards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbbf61b0-5dd2-4643-8857-b7a10ec76ec8",
		Name:         "Living Lies of Loki",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+0 for each other Illusion you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				// "Illusion" is a word of the source's text (CR 612.2).
				illusion := source.TextSubtype("Illusion")
				c.Power += permanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != source.InstanceID && other.IsCreature() && other.HasSubtype(illusion)
				})
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Living Lies of Loki — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
