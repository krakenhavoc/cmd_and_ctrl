package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Laccolith Rig — Enchantment — Aura {R}:
//
//	"Enchant creature
//	 Whenever enchanted creature becomes blocked, you may have it deal damage
//	 equal to its power to target creature. If you do, the first creature
//	 assigns no combat damage this turn."
//
// #3531: the trigger is the Aura's own, on the EventBecomesBlocked of the
// enchanted creature (CR 509.1h). The Aura's controller targets and chooses;
// the enchanted creature deals the damage with its power as the ability
// resolves and is then marked by the CR 510.1a gate.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0b583980-1508-4247-8a28-7127241d8ec0",
		Name:         "Laccolith Rig",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesBlocked && source.IsAttachedTo(ev.CardID)
			}, "Laccolith Rig — you may have it deal damage equal to its power to target creature", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				host := enchantedCreatureOf(ctx)
				if host == uuid.Nil {
					return nil
				}
				target := firstLegalCardTarget(ctx)
				return MayChoice{
					Question: "Laccolith Rig — have the enchanted creature deal damage equal to its power to the target creature? (if you do, it assigns no combat damage this turn)",
					OnYes: func(ctx *Context) error {
						amount := b31CurrentPowerOf(ctx.Game, host)
						if err := (DealDamage{Source: host, Target: target, Amount: amount}).Apply(ctx); err != nil {
							return err
						}
						enchantedAssignsNoCombatDamage(ctx, host, "Laccolith Rig — assigns no combat damage this turn")
						return nil
					},
				}.Apply(ctx)
			}), TargetCreature("target creature")),
		},
	})
}
