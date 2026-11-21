package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swallow Whole — Sorcery {W} (#3868):
//
//	"As an additional cost to cast this spell, tap an untapped creature
//	 you control.
//	 Exile target tapped creature. Put a +1/+1 counter on the creature
//	 tapped to pay this spell's additional cost."
//
// The additional cost is New Blood's (TapUntappedCost). The creature it
// tapped is on the spell's payment record as the object it was
// (ctx.TappedToPay), so the counter goes on it only while it is still
// that object on the battlefield: one that left and came back is a new
// object (CR 400.7) and gets nothing.
//
// Targets are chosen before costs are paid (CR 601.2c before 601.2h), so
// the creature that pays was untapped when the target was chosen and
// cannot be the target. The whole spell does nothing if its target is
// gone (CR 608.2b), the counter included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "54d4c632-1e7c-42f1-938c-9b4bd9b9a9b9",
		Name:           "Swallow Whole",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.TapUntappedCost(1, "an untapped creature you control", effects.Creature()),
		Targets:        effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ExileFirstLegalCardTarget(ctx.Game, item); err != nil {
				return err
			}
			for _, ref := range ctx.TappedToPay() {
				if info, ok := ctx.Game.PermanentForEffect(ref); !ok || info.Left {
					continue
				}
				if err := (effects.AddCounter{Target: ref.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
