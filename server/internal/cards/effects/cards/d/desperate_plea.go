package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Plea — Sorcery — Lesson {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Choose one or both —
//	 • Return target creature card from your graveyard to the
//	   battlefield if its power is less than or equal to the sacrificed
//	   creature's power.
//	 • Destroy target creature."
//
// The comparison reads the sacrificed creature's power as it last
// existed on the battlefield (CR 608.2h), from the payment record (ADR
// 0113 §1), as a comparison: a negative power is compared as it is (CR
// 107.1b). Targets are chosen before the cost is paid (CR 601.2c before
// 601.2h), so the first mode cannot return the creature just sacrificed
// (the 2025-10-02 ruling).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "bc1ea0ba-46bf-49b6-af95-51eaf1ab915e",
		Name:           "Desperate Plea",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target creature card from your graveyard to the battlefield if its power is less than or equal to the sacrificed creature's power.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()), desperatePleaReturn),
			effects.ModeDoing("Destroy target creature.",
				effects.TargetCreature("target creature"), effects.DestroyTheModesTarget),
		),
	})
}

// desperatePleaReturn is the first mode: the card returns only if its
// power is at most the sacrificed creature's.
func desperatePleaReturn(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	sacrificed, ok := ctx.SacrificedPermanent()
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok || c.PowerForComparison() > sacrificed.Power {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
}
