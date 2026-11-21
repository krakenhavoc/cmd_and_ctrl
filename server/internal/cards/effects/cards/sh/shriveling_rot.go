package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shriveling Rot — Instant {2}{B}{B}:
//
//	"Choose one —
//	 • Until end of turn, whenever a creature is dealt damage, destroy it.
//	 • Until end of turn, whenever a creature dies, that creature's
//	   controller loses life equal to its toughness.
//	 Entwine {2}{B} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). Each bullet creates a delayed triggered
// ability with a stated duration (CR 603.7a-b), so it triggers on every
// matching event this turn and is controlled by the caster (CR 603.7d).
// Both are each-conditions (game.DelayedEachCondition): "it" and "that
// creature" are the instance's payload, recorded with the object's epoch.
//
// The first bullet triggers once for each creature dealt damage at the
// same time, however many sources dealt it (CR 603.2c, 120.4b, 510.2;
// #3997): DelayedOnEvent.OncePerBatch on an each-condition. A creature
// two blockers hit is destroyed once, so one regeneration shield saves
// it. "Destroy it" destroys that object if it is still on the
// battlefield (CR 603.7c): a creature the damage killed, or that left and
// came back, is not affected.
//
// The second reads the creature as it last existed on the battlefield
// (CR 603.10a, 608.2h): its controller then, and its toughness with
// counters and effects. A toughness of 0 or less loses no life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9ce8691a-d9e2-45b2-bda0-2c3891b55307",
		Name:          "Shriveling Rot",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{2}{B}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Until end of turn, whenever a creature is dealt damage, destroy it.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DelayedOnEvent{
						Label:        "Shriveling Rot — destroy the creature that was dealt damage",
						On:           []game.EventKind{game.EventDealDamage},
						Condition:    effects.EachCreatureDealtDamageCondition,
						Body:         effects.DestroyThePayloadPermanentBody,
						Repeats:      true,
						OncePerBatch: true,
					}.Apply(ctx)
				}),
			effects.ModeDoing("Until end of turn, whenever a creature dies, that creature's controller loses life equal to its toughness.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DelayedOnEvent{
						Label:     "Shriveling Rot — that creature's controller loses life equal to its toughness",
						On:        []game.EventKind{game.EventLTB},
						Condition: effects.EachCreatureDiedCondition,
						Body:      effects.DiedControllerLosesLifeEqualToToughnessBody,
						Repeats:   true,
					}.Apply(ctx)
				}),
		)),
	})
}
