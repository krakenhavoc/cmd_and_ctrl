package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fight with Fire — Sorcery {2}{R} (Dominaria):
//
//	"Kicker {5}{R} (You may pay an additional {5}{R} as you cast this spell.)
//	 Fight with Fire deals 5 damage to target creature. If this spell was
//	 kicked, it deals 10 damage divided as you choose among any number of
//	 targets instead. (Those targets can include players and planeswalkers.)"
//
// Unkicked it is a 5-damage Shock. Kicked, the kicker's own target clause
// replaces the printed one outright ("instead", CR 702.33d): any number of
// targets of any kind, dividing 10 damage with at least 1 each (CR 601.2d),
// chosen as the spell is cast, and each target still legal on resolution takes
// the share announced for it with no redistribution (CR 608.2b).
//
// No simplification.
func init() {
	kicker := effects.Kicker("{5}{R}")
	kicker.Targets = effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(10))
	effects.Register(effects.Spec{
		OracleID:      "202b275e-e0c4-482d-abe9-c5e6360208f3",
		Name:          "Fight with Fire",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target creature"),
		Purpose:       effects.ForTargets(effects.DamageToTarget(0, 5)),
		OptionalCosts: []game.AdditionalCost{kicker},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.WasKicked() {
				return effects.DealDividedDamage(ctx)
			}
			return effects.OldB208DamageClauseTarget(ctx, 0, 5)
		},
	})
}
