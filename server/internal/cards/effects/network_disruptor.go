package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "0d64d150-1716-4da4-9cfe-ee7749b89918",
		Name:            "Network Disruptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Network Disruptor — tap target permanent", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id := firstLegalCardTarget(ctx); id != uuid.Nil {
					return TapTarget{Target: id}.Apply(ctx)
				}
				return nil
			}), TargetPermanent("target permanent")),
		},
	})
}
