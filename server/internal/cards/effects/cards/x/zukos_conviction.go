package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuko's Conviction — Instant {B}:
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 Return target creature card from your graveyard to your hand. If
//	 this spell was kicked, instead put that card onto the battlefield
//	 tapped."
//
// One target, two destinations: unkicked it goes to its owner's hand,
// kicked it enters tapped under its owner's control (the caster, since
// it came from the caster's graveyard). The kicker is the engine's
// optional additional cost, read with ctx.WasKicked(). A target that
// has left the graveyard makes the spell do nothing (CR 608.2b).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f034e8a3-82f7-4c70-93ea-0bf2688f66e1",
		Name:          "Zuko's Conviction",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		Targets:       effects.TargetCreatureInYourGraveyard(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if ctx.WasKicked() {
				return effects.ReturnFromGraveyard{Target: target, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
			}
			return effects.ReturnFromGraveyard{Target: target, Dest: game.ZoneHand}.Apply(ctx)
		},
	})
}
