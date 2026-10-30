package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cut Propulsion — Instant {2}{R}:
//
//	"Target creature deals damage to itself equal to its power. If that
//	 creature has flying, it deals twice that much damage to itself
//	 instead."
//
// One target clause and one DealDamage whose source and recipient are the
// same permanent. The CREATURE is the damage source, not the spell, which is
// what makes the rest of the card work: a lifelink creature gains its
// controller that much life, a deathtouch creature kills itself with any
// positive amount (CR 702.2b, CR 702.15b), protection from the creature
// itself is not a way to dodge it (the spell targeted it, not the damage),
// and a prevention effect on the creature or on its source applies.
//
// Both numbers are read as the spell resolves, never when it was cast: a
// pump in response makes the creature hit itself harder, and a creature that
// lost flying in response deals its power once rather than twice
// (CR 608.2h). A creature that left the battlefield is an illegal target and
// the spell does not resolve (CR 608.2b), so there is no last-known reading
// to make. A creature with power 0 or less deals nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf3234f2-09d7-4d8b-b871-3bf0fcb75b6d",
		Name:         "Cut Propulsion",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    cutPropulsionResolve,
	})
}

// cutPropulsionResolve is the spell's whole resolution: the target creature
// deals damage to itself equal to its power, or twice that if it has flying.
func cutPropulsionResolve(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	amount := c.CurrentPower()
	if game.HasKeyword(&c, "flying") {
		amount *= 2
	}
	return DealDamage{Source: t.ID, Target: t.ID, Amount: amount}.Apply(ctx)
}
