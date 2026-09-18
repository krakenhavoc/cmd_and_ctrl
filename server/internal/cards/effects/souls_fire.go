package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul's Fire — Instant {2}{R} (EDHREC rank 2589):
//
//	"Target creature you control deals damage equal to its power to
//	 any target."
//
// The red "bite" at instant speed. Two target CLAUSES (#764) read
// positionally: slot 0 is the damage SOURCE, whose power is read as
// the spell resolves (a pump in response scales it) and whose colour,
// deathtouch and lifelink are what the damage carries, as printed;
// slot 1 is the recipient, any target.
//
// Each clause carries its own predicate, so "a creature you control"
// is refused at announce (CR 601.2c) rather than discovered at
// resolution, and each slot is re-checked against its own clause at
// resolution (CR 608.2b): a creature that left in response deals
// nothing, a recipient that left takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62d7ed6e-c386-477e-b155-982c3790f842",
		Name:         "Soul's Fire",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetAny(),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			attacker, ok := ctx.ClauseTarget(0)
			if !ok || attacker.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(attacker.ID)
			if !ok {
				return nil
			}
			return DealDamage{Source: attacker.ID, Target: victim.ID, Amount: c.CurrentPower()}.Apply(ctx)
		},
	})
}
