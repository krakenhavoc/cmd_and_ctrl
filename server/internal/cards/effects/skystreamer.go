package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skystreamer — Creature — Griffin {4}{W}:
//
//	"Flying
//	 Assist (Another player can pay up to {4} of this spell's cost.)
//	 When this creature enters, target player gains 4 life."
//
// The ETB is a plain targeted life gain on any player: "target player"
// is any seated, non-eliminated player, and the gain goes to whichever
// player the pick names.
func init() {
	Register(Spec{
		OracleID:        "aaf6d66b-67b1-43f1-a8bf-db83644f7d04",
		Name:            "Skystreamer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist, "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Skystreamer — target player gains 4 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return GainLife{Player: t.ID, Amount: 4}.Apply(ctx)
							}
						}
						return nil
					}),
				TargetPlayer("target player")),
		},
	})
}
