package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swift Demise — {2}{B} Instant (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Swift Demise deals 1 damage to target creature. Then destroy each
//	 creature you don't control that was dealt damage this turn."
//
// The ping lands first, so the target counts if it is one you don't
// control and survived it; then every creature an opponent controls that
// the turn's events show was dealt damage since it last entered the
// battlefield (DealtDamageThisTurn, Covert Cutpurse's predicate) is
// destroyed. Your own creatures are spared, damaged or not. With the
// target gone before resolution the spell does not resolve at all
// (CR 608.2b), so nothing is destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7eaf3fd0-e141-42c2-874d-b471f4a94ad2",
		Name:         "Swift Demise",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Targets: game.ForTargets(DamageToTarget(0, 1)),
			Answers: game.AnswerRemove,
			// "each creature you don't control that was dealt damage this
			// turn": opponents' creatures, sparing the undamaged ones.
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true, Partial: true},
		},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}.Apply(ctx)
				}
				return nil
			}),
			DestroyAllMatching{Match: And(Creature(), OpponentControls(), DealtDamageThisTurn())},
		),
	})
}
