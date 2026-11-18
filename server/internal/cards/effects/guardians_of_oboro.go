package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardians of Oboro — Creature — Moonfolk Samurai {2}{U}, 3/4 (Kamigawa:
// Neon Dynasty):
//
//	"Defender
//	 Modified creatures you control can attack as though they didn't have
//	 defender. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// A layer 6 static writing an unrestricted DefenderWaiver on each modified
// creature you control (Bristlepack Sentry's seam), read live at declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "42b28dfb-de4d-4005-b0d7-1db3917cf159",
		Name:            "Guardians of Oboro",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{{
			Label: "Modified creatures you control can attack as though they didn't have defender.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && b07IsModified(g, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				w := game.DefenderWaiver{}
				if source != nil {
					w.Source, w.SourceName = source.InstanceID, source.Name
				}
				c.DefenderWaivers = append(c.DefenderWaivers, w)
			},
		}},
	})
}
