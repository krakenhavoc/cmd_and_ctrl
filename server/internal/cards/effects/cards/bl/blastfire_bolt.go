package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blastfire Bolt — Instant {5}{R} (Magic 2015, #5556):
//
//	"Blastfire Bolt deals 5 damage to target creature. Destroy all Equipment
//	 attached to that creature."
//
// The damage, then the destruction, in printed order: the creature has not
// yet died to the damage (state-based actions wait for the resolution), so
// the Equipment is still attached to it. A target that is illegal as the
// spell resolves leaves the whole spell without effect (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcddeac9-8de7-4347-a956-522f4f2fc63b",
		Name:         "Blastfire Bolt",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			return effects.S82OldB359DestroyEquipmentAttachedTo(ctx, t.ID)
		},
	})
}
