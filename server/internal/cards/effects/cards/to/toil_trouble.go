package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toil // Trouble — split card, oracle 97759579-b676-46e8-b566-fcad10751920
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Toil    — Sorcery {2}{B}: "Target player draws two cards and loses 2 life."
//	Trouble — Sorcery {2}{R}: "Trouble deals damage to target player equal to
//	          the number of cards in that player's hand."
//	          Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Toil under the bare oracle ID, Trouble under
// "#1"); the fused cast is the engine's (CR 702.102), Toil resolving first, so a
// fused Trouble counts the two cards Toil just drew. Toil is Blood Pact's draw
// and life loss (a loss, not damage). Trouble reads the target's hand size as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97759579-b676-46e8-b566-fcad10751920",
		Name:         "Toil",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B32TargetPlayerDrawsAndLosesLife(item, ctx, 2, 2)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "97759579-b676-46e8-b566-fcad10751920#1",
		Name:         "Trouble",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			p := ctx.Game.PlayerByIDForEffect(t.ID)
			if p == nil {
				return nil
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: p.Hand.Size()}.Apply(ctx)
		},
	})
}
