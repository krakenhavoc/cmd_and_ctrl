package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindeye Drake — Creature — Drake {4}{U}, 2/5 (Gatecrash, #5565):
//
//	"Flying
//	 When this creature dies, target player mills five cards."
//
// One trigger on its own death with a player target chosen as it goes on the
// stack (CR 603.3d); a player that has left by resolution is an illegal
// target and nothing is milled (CR 608.2b). A library of fewer than five
// cards mills what it has (CR 701.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e33ec9a-e148-4aa6-bff8-b356c45fb369",
		Name:            "Mindeye Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Mindeye Drake — target player mills five cards",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return effects.MillCards{Player: t.ID, N: 5}.Apply(ctx)
							}
						}
						return nil
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
