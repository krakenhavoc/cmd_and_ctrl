package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lithomancer's Focus — Instant {W}:
//
//	"Target creature gets +2/+2 until end of turn. Prevent all damage
//	 that would be dealt to that creature this turn by colorless
//	 sources."
//
// The shield protects the target and reads #2026's colourless test (CR
// 105.2c) as each source would deal damage (CR 609.7b; the ruling: it
// prevents damage "from any colorless source that turn, even if that
// source didn't exist or wasn't colorless as Lithomancer's Focus
// resolved"). Combat damage or not.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8a27a87-fc4c-44b5-9ba8-505bbf42164e",
		Name:         "Lithomancer's Focus",
		Purpose:      effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump|game.AnswerPrevent),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTheTargetUntilEOT(2, 2, "Lithomancer's Focus — +2/+2")(item, ctx); err != nil {
				return err
			}
			return effects.PreventDamageFromSource{Protect: effects.ShieldTheTarget, Filter: game.DamageSourceFilter{Colorless: true}}.Apply(ctx)
		},
	})
}
