package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wear // Tear — split card, oracle 9842734c-1eac-4509-a731-4c22017ae586:
//
//	Wear — Instant {1}{R}: "Destroy target artifact."
//	Tear — Instant {W}: "Destroy target enchantment."
//	       Fuse (You may cast one or both halves of this card from
//	       your hand.)
//
// ADR 0103 made every split card castable as either half (CR 709.3)
// and gave fuse a cast of its own (CR 702.102), so both halves register
// as ADR 0034 faces — Wear under the bare oracle ID, Tear under "#1" —
// and the fused cast needs nothing more: the engine builds its
// definition from these two entries, with Wear's clause first and
// Tear's second, and resolves Wear then Tear, each destroying its own
// target (game/split_fuse.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9842734c-1eac-4509-a731-4c22017ae586",
		Name:         "Wear",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
	Register(Spec{
		OracleID:     "9842734c-1eac-4509-a731-4c22017ae586#1",
		Name:         "Tear",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyChosenPermanent(ctx.Game, item)
		},
	})
}
