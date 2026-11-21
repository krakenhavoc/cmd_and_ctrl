package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ant-Man's Air Force — Creature — Insect {1}{U}, 2/1 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 Whenever this creature attacks, up to one target creature gets -1/-0
//	 until end of turn."
//
// "Up to one" lets the trigger go on the stack with no target, and a
// target that has left by resolution is skipped (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d7eea87-b60f-4ab9-b485-30aadda0cd79",
		Name:            "Ant-Man's Air Force",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Ant-Man's Air Force — up to one target creature gets -1/-0",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.BoostUntilEOT{Target: t.ID, Power: -1, Label: "Ant-Man's Air Force — -1/-0"}.Apply(ctx)
						}
					}
					return nil
				}), effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
