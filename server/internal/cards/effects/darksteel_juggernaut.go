package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Juggernaut — Artifact Creature — Juggernaut {5}, */*:
//
//	"Indestructible
//	 Darksteel Juggernaut's power and toughness are each equal to the
//	 number of artifacts you control.
//	 This creature attacks each combat if able."
//
// Crusader of Odric's layer 7a characteristic-defining P/T, counting
// artifacts the controller has (Juggernaut itself is one, so it is never
// below 1/1 on the battlefield). Juggernaut's attacks-each-combat static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4893a301-29a8-4a0e-845c-8ddb5cb8db25",
		Name:            "Darksteel Juggernaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{
			{
				Label:     "Darksteel Juggernaut's power and toughness are each equal to the number of artifacts you control.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7A_CDA,
				AppliesTo: selfOnly,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := countControlled(g, source.Controller, func(o game.Card) bool { return o.IsArtifact() })
					c.Power, c.Toughness = n, n
				},
			},
			AttacksEachCombat(),
		},
	})
}
