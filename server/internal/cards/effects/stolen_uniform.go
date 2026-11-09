package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stolen Uniform — Instant {U} (Final Fantasy, #4462):
//
//	"Choose target creature you control and target Equipment. Gain control
//	 of that Equipment until end of turn. Attach it to the chosen creature.
//	 When you lose control of that Equipment this turn, if it's attached to
//	 a creature you control, unattach it."
//
// Two target clauses, in order: the creature, then the Equipment. The
// control change is a layer-2 effect until end of turn (GainControl); the
// attach follows it, and only when the creature is still a legal target — a
// creature that has become illegal still lets the Equipment change hands but
// is not attached to (the rulings). The last sentence is a delayed trigger
// (CR 603.7) watching EventControlChanged for that Equipment going from the
// controller to someone else, whether the until-end-of-turn duration ended or
// another effect took it, and it is an intervening "if" (CR 603.4): checked
// as it triggers and again as it resolves, the Equipment still attached to a
// creature the controller controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17c0cf45-a031-40e8-a37f-eda98a8f785b",
		Name:         "Stolen Uniform",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetPermanent("target Equipment", HasSubtype("Equipment")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			creature, haveCreature := ctx.ClauseTarget(0)
			equipment, haveEquipment := ctx.ClauseTarget(1)
			if !haveEquipment {
				return nil
			}
			if err := (GainControl{
				Target: equipment.ID, Controller: item.Controller,
				Label: "Stolen Uniform — gain control of the Equipment until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if haveCreature {
				if err := ctx.Game.AttachForEffect(equipment.ID, game.TargetRef{Kind: game.TargetCard, ID: creature.ID}); err != nil {
					return err
				}
			}
			return DelayedOnEvent{
				Label:     "Stolen Uniform — unattach the Equipment",
				On:        []game.EventKind{game.EventControlChanged},
				Condition: stolenUniformLostControl,
				Cards:     []uuid.UUID{equipment.ID},
				Body:      stolenUniformUnattachBody,
				// The theft ends in the cleanup step, which is still "this turn".
				ThroughCleanup: true,
			}.Apply(ctx)
		},
	})
}

// stolenUniformLostControl is "when you lose control of that Equipment",
// with its intervening "if it's attached to a creature you control" (CR
// 603.4): the Equipment changed controller away from the trigger's
// controller, and is attached to a creature that controller controls.
var stolenUniformLostControl = game.DelayedCondition("stolen-uniform/lost-control-of-the-equipment",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventControlChanged || ev.Target != dt.Controller || len(dt.Cards) == 0 || ev.CardID != dt.Cards[0] {
			return false
		}
		return equipmentAttachedToCreatureControlledBy(g, dt.Cards[0], dt.Controller)
	})

// stolenUniformUnattachBody unattaches the Equipment if it is still attached
// to a creature its stealer controls, rechecked as the trigger resolves.
var stolenUniformUnattachBody = game.SimpleDelayedBody("stolen-uniform/unattach-the-equipment",
	func(g *game.Game, item *game.StackItem) error {
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard || !equipmentAttachedToCreatureControlledBy(g, t.ID, item.Controller) {
				continue
			}
			if err := g.UnattachForEffect(t.ID); err != nil {
				return err
			}
		}
		return nil
	})
