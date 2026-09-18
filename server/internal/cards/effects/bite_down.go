package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bite Down — Instant {1}{G} (EDHREC rank 2758):
//
//	"Target creature you control deals damage equal to its power to
//	 target creature or planeswalker you don't control."
//
// The green one-sided fight. Two target CLAUSES (#764), read
// positionally — the Soul's Fire shape: slot 0 is the damage SOURCE,
// whose power is read as the spell resolves (a pump in response
// scales it) and whose deathtouch and lifelink the damage carries,
// as printed; slot 1 is the recipient. A creature that left in
// response deals nothing; a recipient that left takes nothing.
//
// Each slot carries its OWN predicate, so "you control" and "you
// don't control" are enforced at announce (CR 601.2c) and re-checked
// per slot at resolution (CR 608.2b). Before #764 a target spec was
// one predicate over every slot, so the clause had to widen to "a
// creature or planeswalker" over both halves and check the rest at
// resolution, where a mismatched pair simply did nothing; that
// caveat is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "623903de-3c04-4745-9af3-d7ec9fb2574d",
		Name:         "Bite Down",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetPermanent("target creature or planeswalker you don't control",
				Or(Creature(), Planeswalker()), OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok {
				return nil
			}
			return DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
		},
	})
}
