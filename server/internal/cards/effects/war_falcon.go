package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War Falcon — Creature — Bird {W}, 2/1 (Magic 2013, #5568):
//
//	"Flying
//	 This creature can't attack unless you control a Knight or a Soldier."
//
// Mouser Mark III's attack gate with a creature-type clause: any
// permanent you control with the Knight or Soldier subtype opens it,
// the Falcon's own types included (it has neither).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "406bccb6-6602-4486-a556-e0b550650253",
		Name:            "War Falcon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label: "This creature can't attack unless you control a Knight or a Soldier.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					len(permanentsControlledByMatching(g, source.Controller, Or(Subtype("Knight"), Subtype("Soldier")))) == 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
	})
}
