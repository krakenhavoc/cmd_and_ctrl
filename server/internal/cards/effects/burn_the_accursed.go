package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burn the Accursed — Instant {4}{R}:
//
//	"Burn the Accursed deals 5 damage to target creature and 2 damage to
//	 that creature's controller. If that creature would die this turn,
//	 exile it instead."
//
// One target. "That creature's controller" is read as the spell
// resolves, before the damage, while the creature is on the battlefield
// (CR 608.2h) — Minion Missile's reading. A target that has become
// illegal counters the spell (CR 608.2b), so the controller is not dealt
// the 2 either. The replacement is the spell's (ADR 0108 §1): the
// creature is marked whether or not its damage was dealt.
//
// The 5 and the 2 are one damage instance (one printed instruction, CR
// 615.8, ADR 0108 PR 0), dealt one event after the other as every
// multi-recipient instruction in the engine is.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b5aae42b-3fde-4f10-b85e-882c528badef",
		Name:         "Burn the Accursed",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			// One sentence, one damage instance (CR 615.8, ADR 0108 PR 0).
			err := ctx.Game.DamageInstanceForEffect(func() error {
				if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
				return DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 2}.Apply(ctx)
			})
			if err != nil {
				return err
			}
			return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
		},
	})
}
