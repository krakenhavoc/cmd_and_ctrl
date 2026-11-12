package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cleansing Beam — Instant {4}{R}:
//
//	"Radiance — Cleansing Beam deals 2 damage to target creature and each
//	 other creature that shares a color with it."
//
// The creatures are read as the spell resolves (radianceSet, #5736),
// whoever controls them, and take 2 damage each as one instruction
// (radianceDamage). A colourless target is the only creature hit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c208258-f987-4d41-8eee-62e188d47ffb",
		Name:         "Cleansing Beam",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(DamageToTarget(0, 2))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return radianceDamage(ctx, 2, game.KeyedTail{})
		},
	})
}
