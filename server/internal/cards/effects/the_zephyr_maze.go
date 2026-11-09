package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Zephyr Maze — Plane — Kyneth (Planechase Anthology, #4870):
//
//	"Creatures with flying get +2/+0.
//	 Creatures without flying get -2/-0.
//	 Whenever chaos ensues, target creature gains flying until end of turn."
//
// A plane (ADR 0216 §6). Both statics are layer 7c, which runs after the
// ability layer, so a creature that has gained or lost flying is judged
// by the flying it has then (CR 613.1f before 613.4c). The chaos ability
// grants flying until end of turn (layer 6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d84be6d-49a5-4200-8cd3-d8653dfd0899",
		Name:         "The Zephyr Maze",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			zephyrMazeStatic("Creatures with flying get +2/+0.", true, 2),
			zephyrMazeStatic("Creatures without flying get -2/-0.", false, -2),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("The Zephyr Maze — target creature gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "The Zephyr Maze — flying"}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}

func zephyrMazeStatic(label string, flying bool, power int) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if planeHasFlying(c.Abilities) == flying {
				c.Power += power
			}
		},
	}
}

func planeHasFlying(abilities []string) bool {
	for _, a := range abilities {
		if a == "flying" {
			return true
		}
	}
	return false
}
