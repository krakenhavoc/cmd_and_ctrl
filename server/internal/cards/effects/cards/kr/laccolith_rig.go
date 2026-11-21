package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "0b583980-1508-4247-8a28-7127241d8ec0",
		Name:         "Laccolith Rig",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesBlocked && source.IsAttachedTo(ev.CardID)
			}, "Laccolith Rig — you may have it deal damage equal to its power to target creature", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				host := effects.EnchantedCreatureOf(ctx)
				if host == uuid.Nil {
					return nil
				}
				target := effects.FirstLegalCardTarget(ctx)
				return effects.MayChoice{
					Question: "Laccolith Rig — have the enchanted creature deal damage equal to its power to the target creature? (if you do, it assigns no combat damage this turn)",
					Then:     laccolithRigAnswer,
					Carry:    []uuid.UUID{host, target},
				}.Apply(ctx)
			}), effects.TargetCreature("target creature")),
		},
	})
}

// laccolithRigAnswer is the "yes": the enchanted creature (the first carried
// ID) deals damage equal to its power to the target (the second), then
// assigns no combat damage this turn. Keyed (#4291).
var laccolithRigAnswer = effects.MayYes("confirm/laccolith-rig-damage", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	host, target := effects.CarryAt(r, 0), effects.CarryAt(r, 1)
	amount := effects.B31CurrentPowerOf(ctx.Game, host)
	if err := (effects.DealDamage{Source: host, Target: target, Amount: amount}).Apply(ctx); err != nil {
		return err
	}
	effects.EnchantedAssignsNoCombatDamage(ctx, host, "Laccolith Rig — assigns no combat damage this turn")
	return nil
})
