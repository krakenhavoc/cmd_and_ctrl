package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turn to Slag — Sorcery {3}{R}{R} (Core Set 2021, #5511):
//
//	"Turn to Slag deals 5 damage to target creature. Destroy all Equipment
//	 attached to that creature."
//
// Damage first, then every Equipment attached to the creature as the second
// sentence is read (CR 608.2c). With the target illegal the spell does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4339f621-8502-4433-a37a-654591ca1b43",
		Name:         "Turn to Slag",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			id := legal[0].ID
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			return effects.S82OldB109DestroyAttachedEquipment(ctx, id)
		},
	})
}
