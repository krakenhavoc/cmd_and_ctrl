package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Network Disruptor — Artifact Creature — Moonfolk Rogue {U}, 1/1
// (Kamigawa: Neon Dynasty, S82):
//
//	"Flying
//	 When this creature enters, tap target permanent."
//
// Any permanent, one's own included; a target that left is not tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d64d150-1716-4da4-9cfe-ee7749b89918",
		Name:            "Network Disruptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Network Disruptor — tap target permanent", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id := effects.FirstLegalCardTarget(ctx); id != uuid.Nil {
					return effects.TapTarget{Target: id}.Apply(ctx)
				}
				return nil
			}), effects.TargetPermanent("target permanent")),
		},
	})
}
