package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Showstopping Surprise — Instant {3}{R}{R} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Choose target creature you control. Turn it face up if it's face down.
//	 Then it deals damage equal to its power to each other creature."
//
// The turn is the free effect (CR 708.8, a no-op when it is already face
// up); the power is read afterwards, so a creature that turned up bigger
// deals its real power, as one damage event from that creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e16cbcbd-b03a-404c-8d1e-0a05848181d9",
		Name:         "Showstopping Surprise",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if c, found := ctx.Game.LookupCardForEffect(id); found && c.FaceDownIsPermanent() {
					return TurnFaceUp{Targets: []uuid.UUID{id}}.Apply(ctx)
				}
				return nil
			}),
			Instruction(targetDealsItsPowerToEachOtherCreature),
		),
	})
}
