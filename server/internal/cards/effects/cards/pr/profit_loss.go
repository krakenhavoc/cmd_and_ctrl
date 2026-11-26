package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Profit // Loss — split card, oracle c41e8a99-e4a8-4323-a59d-265266e29fa9
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Profit — Instant {1}{W}: "Creatures you control get +1/+1 until end of turn."
//	Loss   — Instant {2}{B}: "Creatures your opponents control get -1/-1 until
//	         end of turn."
//	         Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Profit under the bare oracle ID, Loss under
// "#1"); the fused cast is the engine's (CR 702.102). Each is a layer-7
// modification over the creatures on the battlefield as it resolves, until
// cleanup (CR 611.2c): Biting Rain's sweep shape, scoped to your creatures and to
// your opponents'.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c41e8a99-e4a8-4323-a59d-265266e29fa9",
		Name:         "Profit",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.BoostUntilEOT{
			Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
			Label: "Profit — creatures you control get +1/+1 until end of turn",
		}),
	})
	effects.Register(effects.Spec{
		OracleID:     "c41e8a99-e4a8-4323-a59d-265266e29fa9#1",
		Name:         "Loss",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve: effects.ResolveSteps(effects.BoostUntilEOT{
			Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -1, Toughness: -1,
			Label: "Loss — creatures your opponents control get -1/-1 until end of turn",
		}),
	})
}
