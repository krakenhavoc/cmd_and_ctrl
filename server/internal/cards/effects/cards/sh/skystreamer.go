package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "aaf6d66b-67b1-43f1-a8bf-db83644f7d04",
		Name:            "Skystreamer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Skystreamer — target player gains 4 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return effects.GainLife{Player: t.ID, Amount: 4}.Apply(ctx)
							}
						}
						return nil
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
