package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bile Urchin — Creature — Spirit {B}, 1/1 (Betrayers of Kamigawa,
// #5602):
//
//	"Sacrifice this creature: Target player loses 1 life."
//
// Any player, its controller included. Life loss, not damage, so it can't
// be prevented.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d406aa7-636a-45c1-903f-11c2ce2ef3e3",
		Name:         "Bile Urchin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Target player loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    SacrificeThis(),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return GainLife{Player: t.ID, Amount: -1}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
