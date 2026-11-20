package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razorgrass Ambush // Razorgrass Field — Modal DFC. Front: Instant
// {1}{W}: "Razorgrass Ambush deals 3 damage to target attacking or
// blocking creature." Back: Land, pay 3 life or enter tapped
// (mdfc_lands.go).
//
// The target clause is judged at announce and again at resolution, so a
// creature that has been removed from combat in response is illegal and
// the spell does nothing. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5da954fa-9001-4557-825c-1462035d21ed",
		Name:         "Razorgrass Ambush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 3}.Apply(ctx)
		},
	})
}
