package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulshriek — Instant {B} (Mirage, S82 OLD b798, #5633):
//
//	"Target creature you control gets +X/+0 until end of turn, where X is the
//	 number of creature cards in your graveyard. Sacrifice that creature at
//	 the beginning of the next end step."
//
// X is counted as the spell resolves (CR 611.2c) and the boost is a layer-7
// change ending at cleanup. The sacrifice is a delayed trigger naming that
// creature (Will Riker, First in the Field's object-pinned body): it asks for the
// object (CR 400.7) still on the battlefield under your control at the end
// step, so a creature that left, came back or changed hands is not sacrificed. The next end step may be this turn's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77ae796e-fc56-47b9-b8a0-97e24fbd60f7",
		Name:         "Soulshriek",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
				return nil
			}
			x := effects.CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
			if x > 0 {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: x, Label: "Soulshriek — +X/+0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			ref, ok := ctx.Game.PermanentRefForEffect(t.ID)
			if !ok {
				return nil
			}
			return effects.ScheduleDelayedTrigger{
				Label:  "Soulshriek — sacrifice that creature",
				Cards:  []uuid.UUID{t.ID},
				Body:   effects.SacrificeListedObjectsYouControlBody,
				Params: game.EffectParams{Objects: []game.ObjectRef{ref}},
			}.Apply(ctx)
		},
	})
}
