package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oath of Kaya — Legendary Enchantment {1}{W}{B} (War of the Spark, #5520):
//
//	"When Oath of Kaya enters, it deals 3 damage to any target and you gain
//	 3 life.
//	 Whenever an opponent attacks a planeswalker you control with one or more
//	 creatures, Oath of Kaya deals 2 damage to that player and you gain
//	 2 life."
//
// The second ability is one trigger per planeswalker attacked in a
// declaration, however many creatures attack it (CR 603.2c): the batch key is
// the attacked planeswalker. The damage goes to the attacking player, and the
// life is gained even when that damage is prevented.
//
// No simplification.
func init() {
	const attackLabel = "Oath of Kaya — 2 damage to that player and you gain 2 life"
	attack := effects.OncePerBatch(effects.On(game.EventAttack,
		func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil || ev.Actor == source.Controller ||
				g.PlayerByIDForEffect(ev.Actor) == nil || ev.Target == uuid.Nil {
				return false
			}
			if g.ClassifyAttackTargetForEffect(ev.Target) != game.AttackTargetPlaneswalker {
				return false
			}
			pw, ok := g.LookupCardForEffect(ev.Target)
			return ok && pw.Controller == source.Controller
		},
		attackLabel,
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if item.Trigger == nil {
				return nil
			}
			who := item.Trigger.Event.Actor
			if err := (effects.DealDamage{Source: ctx.Source(), Target: who, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		}))
	attack.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string {
		return ev.Actor.String() + "/" + ev.Target.String()
	}
	effects.Register(effects.Spec{
		OracleID:     "8b3febb9-0597-46af-b005-982f1a31c83e",
		Name:         "Oath of Kaya",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(
					effects.WhenThisEnters("Oath of Kaya — 3 damage to any target and you gain 3 life",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							if t, ok := ctx.ClauseTarget(0); ok {
								if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
									return err
								}
							}
							return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
						}),
					effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 3))),
			attack,
		},
	})
}
