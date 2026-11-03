package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subterranean Tremors — Sorcery {X}{R} (Secrets of Strixhaven Mystical
// Archive, #4834):
//
//	"Subterranean Tremors deals X damage to each creature without flying.
//	 If X is 4 or more, destroy all artifacts. If X is 8 or more, create
//	 an 8/8 red Lizard creature token."
//
// Earthquake's creature sweep (a damage instance, "without flying" read
// post-layer, so a creature that has lost or gained flying is judged as it
// is), then the two thresholds, each its own step behind the one before
// (CR 608.2c): all artifacts are destroyed at X of 4 or more, so an
// indestructible one stays, and the 8/8 Lizard arrives at X of 8 or more.
// The thresholds read the announced X, not the damage that was actually
// dealt. Nobody takes damage but creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "f0881756-de74-43b1-a9fe-c2fde93faf5a",
		Name:     "Subterranean Tremors",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				return ctx.Game.DamageInstanceForEffect(func() error {
					return damageEachMatching(ctx, And(Creature(), WithoutKeyword("flying")), x)
				})
			}),
			Instruction(func(ctx *Context) error {
				if ctx.X() < 4 {
					return nil
				}
				return DestroyAllMatching{Match: Artifact()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if ctx.X() < 8 {
					return nil
				}
				return CreateToken{Template: TokenCard("8/8 red Lizard"), N: 1}.Apply(ctx)
			}),
		),
	})
}
