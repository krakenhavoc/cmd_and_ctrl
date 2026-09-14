package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bite Down — Instant {1}{G} (EDHREC rank 2758):
//
//	"Target creature you control deals damage equal to its power to
//	 target creature or planeswalker you don't control."
//
// The green one-sided fight. Two target slots read positionally —
// the Soul's Fire shape: slot 0 is the damage SOURCE, whose power is
// read as the spell resolves (a pump in response scales it) and whose
// deathtouch and lifelink the damage carries, as printed; slot 1 is
// the recipient. A creature that left in response deals nothing; a
// recipient that left takes nothing.
//
// DECLARED SIMPLIFICATION, weaker than printed — the Soul's Fire /
// Run Away Together posture: the two slots have different clauses
// ("a creature you control", "a creature or planeswalker you don't
// control"), and a target spec is one predicate over every slot, so
// the spec is "a creature or planeswalker" over both and the per-slot
// halves are checked at RESOLUTION rather than refused at announce.
// A pair that is not (yours, theirs) makes the spell do nothing.
// Never stronger: no line exists that the printed card forbids.
func init() {
	Register(Spec{
		OracleID:     "623903de-3c04-4745-9af3-d7ec9fb2574d",
		Name:         "Bite Down",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The first target must be a creature you control and the second a creature or planeswalker you don't control, but that isn't checked until the spell resolves — a pair that doesn't fit makes the spell do nothing."},
		Targets:      b26TargetCreatureYouControlThenCreatureOrPlaneswalkerYouDontControl(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) < 2 {
				return nil
			}
			biter, victim := item.Targets[0], item.Targets[1]
			if biter.Kind != game.TargetCard || victim.Kind != game.TargetCard ||
				!ctx.IsTargetLegal(biter) || !ctx.IsTargetLegal(victim) {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok || !c.IsCreature() || c.Controller != ctx.Controller() {
				return nil
			}
			v, ok := ctx.Game.LookupCardForEffect(victim.ID)
			if !ok || v.Controller == ctx.Controller() {
				return nil
			}
			return DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
		},
	})
}
