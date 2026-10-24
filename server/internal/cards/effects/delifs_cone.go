package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Delif's Cone — Artifact {0}:
//
//	"{T}, Sacrifice this artifact: This turn, when target creature you
//	 control attacks and isn't blocked, you may gain life equal to its
//	 power. If you do, it assigns no combat damage this turn."
//
// #3531, ADR 0187 amendment 2026-10-16: the ability creates a delayed
// triggered ability (CR 603.7a) about the creature it targeted, which
// triggers the first time that creature attacks and isn't blocked this
// turn and then ceases to exist (CR 603.7b, 603.7c). The creature is the
// trigger's "it" (thatCreatureAttacksUnblockedCondition, keyed by the
// object, so a creature that left and came back is a new object and
// never fires it). "You may" is asked as the trigger resolves; the life
// is its power then, or its last-known power if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8af33199-d3ad-44fe-8ee0-abbe36a2278b",
		Name:         "Delif's Cone",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: This turn, when target creature you control attacks and isn't blocked, you may gain life equal to its power. If you do, it assigns no combat damage this turn.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return whenThatCreatureAttacksUnblocked(ctx, FirstLegalBattlefieldTarget(ctx),
					"Delif's Cone — you may gain life equal to its power", delifsConeBody)
			},
		}},
	})
}

// delifsCone is the delayed trigger's body.
func delifsCone(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	it := firstPayloadCard(ctx)
	if it == uuid.Nil {
		return nil
	}
	return MayChoice{
		Question: "Delif's Cone — gain life equal to the attacking creature's power? (if you do, it assigns no combat damage this turn)",
		Then:     delifsConeAnswer,
		Carry:    []uuid.UUID{it},
	}.Apply(ctx)
}

// delifsConeAnswer is the "yes": gain life equal to the attacking creature's
// power (the carried ID), and it assigns no combat damage this turn. Keyed
// (#4291).
var delifsConeAnswer = MayYes("confirm/delifs-cone-gain-life", func(ctx *Context, r game.ConfirmAnswer) error {
	it := carryAt(r, 0)
	if err := (GainLife{Player: ctx.Controller(), Amount: powerNowOrLastKnown(ctx.Game, it)}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.AssignsNoCombatDamageThisTurnForEffect(ctx.Source(), it, "Delif's Cone — assigns no combat damage this turn")
	return nil
})
