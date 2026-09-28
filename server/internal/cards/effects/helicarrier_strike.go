package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helicarrier Strike — {W} Instant:
//
//	"Teamwork 2 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 2 or more.)
//	 Helicarrier Strike deals 2 damage to target attacking or blocking
//	 creature. If this spell was cast using teamwork, it deals 4 damage
//	 to that creature instead."
//
// #1703: Teamwork(2), with the "attacking or blocking creature" clause
// built from the new AttackingOrBlocking() predicate (targets.go) —
// Card.AttackingTarget / Card.BlockingTarget, both cleared at combat's
// end (CR 506.4, CR 509.1h have no bearing here; the fields themselves
// are cleared by clearCombatLocked), so the clause is legal only
// during the combat that set them, and the legality is re-checked at
// resolution (CR 608.2b) exactly like any other target. No
// simplification.
func init() {
	Register(Spec{
		OracleID:      "ca3cda24-0ecd-4edf-b1b3-54311ad58a51",
		Name:          "Helicarrier Strike",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(2)},
		Targets:       TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			amount := 2
			if ctx.UsedTeamwork() {
				amount = 4
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}.Apply(ctx)
		},
	})
}
