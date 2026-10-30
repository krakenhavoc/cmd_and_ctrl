package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Self-Destruct — Instant {1}{R}:
//
//	"Target creature you control deals X damage to any other target and X
//	 damage to itself, where X is its power."
//
// Two target clauses (CR 601.2c, ADR 0065 §1), read positionally the way Bite
// Down's are: slot 0 is the creature you control, slot 1 is "any other
// target" (a player, creature, planeswalker or battle). Slot 1 is Distinct,
// so the creature can never be its own second target: that is "other", judged
// at announce and again at resolution (Arc Trail's "any other target").
//
// The creature is the damage source for both hits, so its lifelink and
// deathtouch apply to each, and X is its power as the spell RESOLVES, read
// once for both (a pump in response is felt twice). The two hits are one
// printed instruction and so one damage instance (CR 615.8), as Arc Trail's
// are. Each slot is re-checked on its own (CR 608.2b):
//   - the creature gone: it is an illegal target and deals nothing, so the
//     other target is not hit either;
//   - the other target gone: the creature, still legal, still deals X damage
//     to itself, since that half of the sentence names no illegal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98bf6311-bc9b-4bf8-ac83-c041973b86d0",
		Name:         "Self-Destruct",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			selfDestructOtherTarget(),
		),
		OnResolve: selfDestructResolve,
	})
}

// selfDestructOtherTarget is "any other target": TargetAny with the card's
// label, marked Distinct so it must differ from the creature picked first.
func selfDestructOtherTarget() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any other target"
	return Distinct(spec)
}

// selfDestructResolve is the spell's whole resolution: the creature deals X
// damage to the other target and X damage to itself, X being its power now.
func selfDestructResolve(_ *game.StackItem, ctx *Context) error {
	creature, ok := ctx.ClauseTarget(0)
	if !ok || creature.Kind != game.TargetCard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(creature.ID)
	if !ok {
		return nil
	}
	x := c.CurrentPower()
	if x <= 0 {
		return nil
	}
	other, hasOther := ctx.ClauseTarget(1)
	return ctx.Game.DamageInstanceForEffect(func() error {
		if hasOther {
			if err := (DealDamage{Source: creature.ID, Target: other.ID, Amount: x}).Apply(ctx); err != nil {
				return err
			}
		}
		return DealDamage{Source: creature.ID, Target: creature.ID, Amount: x}.Apply(ctx)
	})
}
