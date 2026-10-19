package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razorgrass Ambush // Razorgrass Field — Modal DFC. Front: Instant
// {1}{W}: "Razorgrass Ambush deals 3 damage to target attacking or
// blocking creature." Back: Land, pay 3 life or enter tapped
// (mdfc_lands.go).
//
// The target clause is judged at announce and again at resolution, so a
// creature that has been removed from combat in response is illegal and
// the spell does nothing. No simplification.
func init() {
	Register(Spec{
		OracleID:     "5da954fa-9001-4557-825c-1462035d21ed",
		Name:         "Razorgrass Ambush",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: id, Amount: 3}.Apply(ctx)
		},
	})
}
