package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zealot il-Vec — Creature — Human Rebel {2}{W}, 1/1:
//
//	"Shadow (This creature can block or be blocked by only creatures with shadow.)
//	 Whenever this creature attacks and isn't blocked, you may have it deal 1 damage to target creature. If you do, prevent all combat damage this creature would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the trigger targets a creature
// (it must, even if you won't use it; its ruling), and "you may" is asked
// as it resolves. On yes the Zealot deals 1 damage to the target, then
// its own combat damage this turn is prevented — even if that 1 damage
// was itself prevented or replaced (its ruling). The shield's source is
// the Zealot while it is still the same object (CR 400.7).
//
// No simplifications.
// zealotIlVecRecord files the creature Zealot il-Vec was aimed at.
const zealotIlVecRecord = "zealot-il-vec/target"

// zealotIlVecYes is the damage and then the shield: the shield waits for a
// prompt the damage asks (a CR 616 order; CR 608.2c, #4155).
var zealotIlVecYes = RegisterThenSteps("then-steps/zealot-il-vec-yes",
	Instruction(func(ctx *Context) error {
		ids := ctx.StepRecord(zealotIlVecRecord).IDs
		if len(ids) == 0 {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: ids[0], Amount: 1}.Apply(ctx)
	}),
	Instruction(shieldAgainstThisCombatDamage),
)

func init() {
	Register(Spec{
		OracleID:        "4107d0e3-608b-4024-b433-40c5d8b63549",
		Name:            "Zealot il-Vec",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenAttacksAndIsNotBlockedEffect("Zealot il-Vec — you may have it deal 1 damage to target creature",
				func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
					ctx := NewContext(g, item)
					var target uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							target = t.ID
						}
					}
					if target == uuid.Nil {
						return nil
					}
					return MayChoice{
						Question: "Zealot il-Vec — deal 1 damage to the target creature? (if you do, its combat damage this turn is prevented)",
						Then:     zealotIlVecAnswer,
						Carry:    []uuid.UUID{target},
					}.Apply(ctx)
				}), TargetCreature("target creature")),
		},
	})
}

// zealotIlVecAnswer is the "yes": deal 1 damage to the target (the carried
// ID), and its combat damage this turn is prevented. Keyed (#4291).
var zealotIlVecAnswer = MayYes("confirm/zealot-il-vec-damage", func(ctx *Context, r game.ConfirmAnswer) error {
	ctx.RecordStep(zealotIlVecRecord, []uuid.UUID{carryAt(r, 0)})
	return zealotIlVecYes.Run(ctx)
})
