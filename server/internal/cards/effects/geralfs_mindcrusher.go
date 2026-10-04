package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geralf's Mindcrusher — Creature — Zombie Horror {4}{U}{U}, 5/5:
//
//	"When this creature enters, target player mills five cards.
//	 Undying"
//
// Undying is PrintedKeywords (#2075); each return mills five more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8fa9455-92ac-46e0-bb34-4175c3c66fee",
		Name:            "Geralf's Mindcrusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Geralf's Mindcrusher — target player mills five cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return MillCards{Player: t.ID, N: 5}.Apply(ctx)
					}
					return nil
				}), TargetPlayer("target player")),
		},
	})
}
