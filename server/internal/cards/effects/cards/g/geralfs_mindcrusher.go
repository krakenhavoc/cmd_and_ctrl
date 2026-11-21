package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geralf's Mindcrusher — Creature — Zombie Horror {4}{U}{U}, 5/5:
//
//	"When this creature enters, target player mills five cards.
//	 Undying"
//
// Undying is PrintedKeywords (#2075); each return mills five more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8fa9455-92ac-46e0-bb34-4175c3c66fee",
		Name:            "Geralf's Mindcrusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Geralf's Mindcrusher — target player mills five cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.MillCards{Player: t.ID, N: 5}.Apply(ctx)
					}
					return nil
				}), effects.TargetPlayer("target player")),
		},
	})
}
