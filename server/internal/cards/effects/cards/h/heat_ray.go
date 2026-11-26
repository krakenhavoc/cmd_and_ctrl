package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heat Ray — Instant {X}{R} (Magic 2015, #5556):
//
//	"Heat Ray deals X damage to target creature."
//
// The X announced as the spell is cast, dealt by the spell (CR 107.3m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76ec76b9-da0f-4b9d-ad0a-d734052a5f2b",
		Name:         "Heat Ray",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.X()}.Apply(ctx)
			}
			return nil
		},
	})
}
