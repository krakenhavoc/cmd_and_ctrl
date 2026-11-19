package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspired Ultimatum — Sorcery {U}{U}{R}{R}{R}{W}{W} (Special Guests, S82 BNS):
//
//	"Target player gains 5 life, Inspired Ultimatum deals 5 damage to any
//	 target, then you draw five cards."
//
// Two target clauses, each rechecked on its own (CR 608.2b): a player who is
// gone gains nothing, a damage target that is gone takes nothing, and the
// caster still draws five. The life is gained by the targeted player, not
// the caster; the draw is the caster's ("you").
//
// No simplification.
func init() {
	purpose := ForTargets(
		game.TargetPurpose{Slot: 0, LifeGain: 5},
		DamageToTarget(1, 5),
	)
	purpose.Draws = 5
	Register(Spec{
		OracleID:     "76ca315a-4979-4fda-8d3f-b7f4e5d3a622",
		Name:         "Inspired Ultimatum",
		Completeness: CompletenessFull,
		Targets:      Clauses(TargetPlayer("target player"), TargetAny()),
		Purpose:      purpose,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if who, ok := ctx.ClauseTarget(0); ok {
				if err := (GainLife{Player: who.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{Player: ctx.Controller(), N: 5}.Apply(ctx)
		},
	})
}
