package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Special Move — {2}{R} Instant (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Choose two —
//	 • Jump Kick — Destroy target artifact.
//	 • Dash Attack — Put two +1/+1 counters on target attacking or
//	   blocking creature you control.
//	 • Foot Toss — Target creature you control deals damage equal to its
//	   power to any other target. Then sacrifice it."
//
// A choose-two spell (CR 700.2), each bullet a ModeDoing reading its own
// target group, run in printed order (CR 608.2c), so a Dash Attack that
// grew the creature Foot Toss throws is read with its counters. Foot Toss
// is the HULK SMASH! bite with "any other target" as its second clause
// (Distinct, so the thrower is never its own target), and the thrower is
// sacrificed afterwards. When the other target has become illegal no
// damage is dealt but the still-legal thrower is sacrificed (CR 608.2b);
// when the thrower itself has become illegal the bullet does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7564afbb-b29e-4234-a3b2-3354c6f5c90a",
		Name:         "Special Move",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Jump Kick — Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
			ModeDoing("Dash Attack — Put two +1/+1 counters on target attacking or blocking creature you control.",
				TargetCreature("target attacking or blocking creature you control", AttackingOrBlocking(), YouControl()),
				PlusOneCountersOnTheModesTarget(2)),
			ModeDoing("Foot Toss — Target creature you control deals damage equal to its power to any other target. Then sacrifice it.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					Distinct(specialMoveAnyOtherTarget()),
				),
				specialMoveFootToss),
		),
	})
}

// specialMoveFootToss is the Foot Toss bullet: slot 0 deals damage equal
// to its power to slot 1, then is sacrificed.
func specialMoveFootToss(_ *game.StackItem, ctx *Context, occ int) error {
	thrower, ok := ModeClauseTarget(ctx, occ, 0)
	if !ok {
		return nil
	}
	if victim, ok := ModeClauseTarget(ctx, occ, 1); ok {
		ctx.Game.RecomputeLayersIfStaleLocked()
		if c, found := ctx.Game.LookupCardForEffect(thrower.ID); found {
			if err := (DealDamage{Source: thrower.ID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return SacrificePermanent{Target: thrower.ID}.Apply(ctx)
}

// specialMoveAnyOtherTarget is Foot Toss's second clause: any target other
// than the creature picked for the first (Distinct).
func specialMoveAnyOtherTarget() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any other target"
	return spec
}
