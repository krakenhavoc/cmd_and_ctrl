package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Borderland Behemoth — Creature — Giant Warrior {5}{R}{R}, 4/4
// (Morningtide, #5592):
//
//	"Trample
//	 This creature gets +4/+4 for each other Giant you control."
//
// Boneclub Berserker's count in layer 7c (CR 613.4c), after layer 4 typed
// the permanents, so a changeling counts; "other" is any Giant permanent
// you control but this one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "378a78d8-9001-4aa5-a66b-2aca27fa8180",
		Name:            "Borderland Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +4/+4 for each other Giant you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				giant := source.TextSubtype("Giant")
				n := permanentsYouControl(g, source, func(other *game.Card) bool {
					return other.InstanceID != source.InstanceID && other.HasSubtype(giant)
				})
				c.Power += 4 * n
				c.Toughness += 4 * n
			},
		}},
	})
}
