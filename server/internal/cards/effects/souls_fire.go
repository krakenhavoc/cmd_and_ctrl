package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul's Fire — Instant {2}{R} (EDHREC rank 2589):
//
//	"Target creature you control deals damage equal to its power to
//	 any target."
//
// The red "bite" at instant speed. Two target slots — the creature,
// then the thing it hits — read positionally (the Arc Trail shape):
// slot 0 is the damage SOURCE, whose power is read as the spell
// resolves (a pump in response scales it) and whose colour,
// deathtouch and lifelink are what the damage carries, as printed;
// slot 1 is the recipient, any target. A creature that left in
// response deals nothing; a recipient that left takes nothing.
//
// DECLARED SIMPLIFICATION, weaker than printed — the Run Away
// Together posture: the two slots have different clauses ("a
// creature you control" and "any target"), and a target spec is one
// predicate over every slot, so the spec is TargetAny over both and
// "a creature you control" in the first slot is checked at
// RESOLUTION rather than refused at announce. A first pick that is
// not a creature the caster controls makes the spell do nothing.
// Never stronger: no line exists that the printed card forbids.
func init() {
	Register(Spec{
		OracleID:     "62d7ed6e-c386-477e-b155-982c3790f842",
		Name:         "Soul's Fire",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The first target must be a creature you control, but that isn't checked until the spell resolves — a first pick that isn't one makes the spell do nothing."},
		Targets:      b24TargetCreatureYouControlThenAnyTarget(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) < 2 {
				return nil
			}
			attacker, victim := item.Targets[0], item.Targets[1]
			if attacker.Kind != game.TargetCard || !ctx.IsTargetLegal(attacker) || !ctx.IsTargetLegal(victim) {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(attacker.ID)
			if !ok || !c.IsCreature() || c.Controller != ctx.Controller() {
				return nil
			}
			return DealDamage{Source: attacker.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
		},
	})
}

// b24TargetCreatureYouControlThenAnyTarget is Soul's Fire's two-slot
// clause: TargetAny's set, two picks, labelled for the order the
// card reads in. The first slot's narrower clause is enforced at
// resolution — see the card comment.
func b24TargetCreatureYouControlThenAnyTarget() *game.TargetSpec {
	spec := TargetAny().WithCount(2, 2)
	spec.Label = "target creature you control, then any target"
	return spec
}
