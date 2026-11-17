package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hail Storm — Instant {1}{G}{G} (Alliances, #5634):
//
//	"Hail Storm deals 2 damage to each attacking creature and 1 damage to
//	 you and each creature you control."
//
// All of it is one damage event (DamageInstanceForEffect): an attacking
// creature you control is dealt 2 and 1, 3 in all, and state-based
// actions see the whole result at once. Every attacking creature counts,
// whoever controls it, so cast on your own attack it hits your attackers
// too. It declares a damage sweep that spares the creatures that are
// neither attacking nor yours (ADR 0126 §6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73eacb1f-3d35-4ead-bead-f85e6cfc17ab",
		Name:         "Hail Storm",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.DamageInstanceForEffect(func() error {
				if err := damageEachMatching(ctx, AttackingCreature(), 2); err != nil {
					return err
				}
				if err := (DamageController{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return damageEachMatching(ctx, And(Creature(), YouControl()), 1)
			})
		},
	})
}
