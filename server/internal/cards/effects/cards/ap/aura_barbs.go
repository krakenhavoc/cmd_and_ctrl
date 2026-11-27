package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aura Barbs — Instant — Arcane {2}{R} (Betrayers of Kamigawa, #5602):
//
//	"Each enchantment deals 2 damage to its controller, then each Aura
//	 attached to a creature deals 2 damage to the creature it's attached
//	 to."
//
// Two steps in the printed order, each one damage instance (CR 120.3):
// every enchantment on the battlefield, Auras and enchantment creatures
// included, deals 2 to its controller; then every Aura attached to a
// creature deals 2 to that creature. The Auras are read as the second step
// begins, so one the first step left attached to a creature still counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1aa8fdc3-efeb-450d-9540-2ef0efaf270d",
		Name:         "Aura Barbs",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := ctx.Game.DamageInstanceForEffect(func() error {
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if !c.IsEnchantment() {
						continue
					}
					if err := ctx.Game.DealDamageToPlayerForEffect(c.InstanceID, c.Controller, 2); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if !c.IsAura() || !c.IsAttached() || c.AttachedTo.Kind != game.TargetCard {
						continue
					}
					host, ok := ctx.Game.LookupCardForEffect(c.AttachedTo.ID)
					if !ok || !effects.OnBattlefield(ctx.Game, host.InstanceID) || !host.IsCreature() {
						continue
					}
					if err := ctx.Game.DealDamageToCreatureForEffect(c.InstanceID, host.InstanceID, 2); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
