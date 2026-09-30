package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// HULK SMASH! — {1}{R} Instant:
//
//	"Teamwork 4 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 4 or more.)
//	 Choose one. If this spell was cast using teamwork, choose both
//	 instead.
//	 • Destroy target noncreature artifact.
//	 • Target creature you control deals damage equal to its power to
//	   target creature an opponent controls."
//
// #1703: Teamwork(4) is the cost, and "choose both instead" is
// InsteadIf(2, TeamworkUsed) — #1655's announcement-driven count, so a
// cast using teamwork must take both bullets (CR 702.194b). The second
// bullet is a bite with two clauses in one mode, read per slot; the
// damage is the biter's power as the spell resolves, and nothing is
// dealt if either creature has become an illegal target (CR 608.2b).
// Bullets run in printed order (CR 608.2c). No simplification.
func init() {
	Register(Spec{
		OracleID:      "5f97d49c-d0fa-4776-9455-93c1a172cd83",
		Name:          "HULK SMASH!",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(4)},
		Modes: ChooseOne(
			ModeDoing("Destroy target noncreature artifact.",
				TargetPermanent("target noncreature artifact", Artifact(), Noncreature()),
				DestroyTheModesTarget),
			ModeDoing("Target creature you control deals damage equal to its power to target creature an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				),
				biteTheModesSecondTarget),
		).InsteadIf(2, TeamworkUsed),
	})
}

// biteTheModesSecondTarget is "target creature you control deals
// damage equal to its power to target creature an opponent controls"
// as a modal bullet: slot 0 bites slot 1.
func biteTheModesSecondTarget(_ *game.StackItem, ctx *Context, occ int) error {
	biter, ok := ModeClauseTarget(ctx, occ, 0)
	if !ok {
		return nil
	}
	victim, ok := ModeClauseTarget(ctx, occ, 1)
	if !ok {
		return nil
	}
	ctx.Game.RecomputeLayersIfStaleLocked()
	c, found := ctx.Game.LookupCardForEffect(biter.ID)
	if !found {
		return nil
	}
	return DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
}
