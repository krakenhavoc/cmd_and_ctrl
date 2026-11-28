package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon Mask — Artifact {3} (Visions, S82 older sets batch 789, #5631):
//
//	"{3}, {T}: Target creature you control gets +2/+2 until end of turn.
//	 Return it to its owner's hand at the beginning of the next end step.
//	 (Return it only if it's on the battlefield.)"
//
// The pump, then a delayed trigger (CR 603.7) that returns that object, and
// only that object (CR 400.7), to its owner's hand at the next end step.
//
// No simplification.
func init() {
	const label = "Dragon Mask — +2/+2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "fce29dd8-23b1-465c-a977-526f716dc8f6",
		Name:         "Dragon Mask",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature you control gets +2/+2 until end of turn. Return it to its owner's hand at the beginning of the next end step.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return effects.ScheduleDelayedTrigger{
					At:     game.StepEnd,
					Label:  "Dragon Mask — return " + c.Name + " to its owner's hand",
					Cards:  []uuid.UUID{id},
					Body:   effects.ZaraReturnToHandBody,
					Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
				}.Apply(ctx)
			},
		}},
	})
}
