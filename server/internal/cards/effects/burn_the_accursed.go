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
// No card-specific simplifications. The one engine-wide one, which Arc
// Trail shares: the two damage events are dealt one after the other
// rather than at once (game.DealDamageEachThenForEffect says why), which
// nothing on the board can tell unless a damage replacement pauses for a
// choice between them.
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
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			if err := (DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return ExileIfItWouldDieThisTurn{Target: id}.Apply(ctx)
		},
	})
}
