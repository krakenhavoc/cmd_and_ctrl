package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Momo's Heist — Sorcery {2}{R} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Gain control of target artifact. Untap it. It gains haste. At the
//	 beginning of the next end step, sacrifice it."
//
// Caught Red-Handed's steal, untap and haste, with neither the control
// nor the haste ending at end of turn: the printed card gives them no
// duration. The sacrifice is a delayed trigger at the beginning of the
// next end step, controlled by the caster; it sacrifices the artifact
// only if the caster still controls it then (you can't sacrifice what
// you don't control).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8f05217-1f61-4d6d-be17-7fb15f1d1e32",
		Name:         "Momo's Heist",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (effects.GainControl{
				Target:     target,
				Controller: item.Controller,
				Duration:   game.IndefiniteDuration(),
				Label:      "Momo's Heist — gain control",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.ScopedEffectFor{Target: target, Mods: []game.Mod{game.AddKeywordsMod("haste")},
				Duration: game.IndefiniteDuration(), Label: "Momo's Heist — it gains haste"}).Apply(ctx); err != nil {
				return err
			}
			return effects.ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: item.Controller,
				Label:      "Momo's Heist — sacrifice the artifact",
				Cards:      []uuid.UUID{target},
				Body:       effects.SacrificeListedCardsBody,
			}.Apply(ctx)
		},
	})
}
