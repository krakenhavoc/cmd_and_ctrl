package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ingenious Thief — {1}{U} Creature — Human Rogue 1/1 (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Flying
//	 When this creature enters, look at target player's hand."
//
// A LOOK (CR 701.20): only the controller sees the hand (LookAtHandForEffect).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73ea2949-5812-478d-8f09-00743ce4d40f",
		Name:            "Ingenious Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Ingenious Thief — look at target player's hand",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if t, ok := effects.FirstLegalPlayerTarget(ctx); ok {
							g.LookAtHandForEffect(ctx.Controller(), t)
						}
						return nil
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
