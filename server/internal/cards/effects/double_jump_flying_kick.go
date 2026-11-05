package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Double Jump // Flying Kick — split card (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	Double Jump — Instant {1}{U}: "Put a flying counter on target
//	  creature you control. Until end of turn, it has base power and
//	  toughness 5/5.
//	  Fuse"
//	Flying Kick — Instant {1}{R}: "Target creature you control deals
//	  damage equal to its power to target creature an opponent controls.
//	  Fuse (You may cast one or both halves of this card from your
//	  hand.)"
//
// Wear // Tear's registration (ADR 0103): Double Jump under the bare oracle
// ID, Flying Kick under "#1", and the fused cast is the engine's
// (CR 702.102, game/split_fuse.go), resolving Double Jump first, so a
// creature that is both halves' target bites for its new base 5 plus its
// counters. The flying counter is a keyword counter (CR 122.1b); the base
// power and toughness are a layer 7b effect until end of turn
// (CR 613.4b), so +1/+1 counters and pumps still apply on top. Flying Kick
// is Bite Down's body (oneSidedBite) over "target creature an opponent
// controls": a biter that left deals nothing, a victim that left takes
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb3072c7-393c-43d9-a2c9-f918bd0fd7be",
		Name:         "Double Jump",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (AddCounter{Target: t.ID, Kind: game.CounterFlying, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return untilEndOfTurn(ctx, t.ID, nil, "Double Jump — base power and toughness 5/5",
				game.SetBasePTMods(5, 5)...)
		},
	})
	Register(Spec{
		OracleID:     "cb3072c7-393c-43d9-a2c9-f918bd0fd7be#1",
		Name:         "Flying Kick",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: oneSidedBite,
	})
}
