package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallifrey Falls // No More — split card, oracle d4424585-9564-4ec2-8267-3f5438e1f29e:
//
//	Gallifrey Falls — Instant {4}{R}{R}: "Gallifrey Falls deals 4 damage
//	       to each creature. If a creature dealt damage this way would
//	       die this turn, exile it instead.
//	       Fuse (You may cast one or both halves of this card from your
//	       hand.)"
//	No More — Instant {2}{W}: "Any number of target creatures you control
//	       phase out.
//	       Fuse (You may cast one or both halves of this card from your
//	       hand.)"
//
// Both halves register as ADR 0034 faces, Gallifrey Falls under the bare
// oracle ID and No More under "#1", and the fused cast is the engine's
// (CR 702.102, game/split_fuse.go): Gallifrey Falls resolves first, then
// No More (CR 702.102d), so a fused cast deals the 4 damage and then
// phases your chosen creatures out before any state-based check, and
// the creatures dealt lethal damage that stayed behind are exiled. Each
// half reads only its own targets. Gallifrey Falls
// marks each creature dealt damage from the damage's continuation (ADR
// 0108 §1 decision 3). No More is Clever Concealment's "any number of
// target" phase-out, all at once (CR 702.26).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d4424585-9564-4ec2-8267-3f5438e1f29e",
		Name:         "Gallifrey Falls",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachExileIfDealtDies(ctx, Creature(), 4, false)
		},
	})
	Register(Spec{
		OracleID:     "d4424585-9564-4ec2-8267-3f5438e1f29e#1",
		Name:         "No More",
		Completeness: CompletenessFull,
		Targets: TargetCreature("any number of target creatures you control",
			YouControl()).WithCount(0, 0),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PhaseOut{Targets: legalTargetCards(item, ctx.Game)}.Apply(ctx)
		},
	})
}
