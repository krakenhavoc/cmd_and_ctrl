package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protect // Serve — split card, oracle 2e0cd5a5-3aee-4b32-873a-e5a40bfba3e9
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Protect — Instant {2}{W}: "Target creature gets +2/+4 until end of turn."
//	Serve   — Instant {1}{U}: "Target creature gets -6/-0 until end of turn."
//	          Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Protect under the bare oracle ID, Serve under
// "#1"); the fused cast is the engine's (CR 702.102): each half takes its own
// target (the same creature may be both) and resolves in order. Each is Giant
// Growth's one layer-7 modification until cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e0cd5a5-3aee-4b32-873a-e5a40bfba3e9",
		Name:         "Protect",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: 2, Toughness: 4, Label: "Protect — +2/+4"}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "2e0cd5a5-3aee-4b32-873a-e5a40bfba3e9#1",
		Name:         "Serve",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: -6, Label: "Serve — -6/-0"}.Apply(ctx)
		},
	})
}
