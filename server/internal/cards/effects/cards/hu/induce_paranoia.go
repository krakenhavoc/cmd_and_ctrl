package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Induce Paranoia — Instant {2}{U}{U} (Ravnica: City of Guilds, S82 OLD b582):
//
//	"Counter target spell. If {B} was spent to cast this spell, that spell's
//	 controller mills X cards, where X is the spell's mana value."
//
// Cancel's counter with River's Grasp's colour reading. The spell's
// controller and mana value (an announced X counted, CR 202.3e) are read
// before the counter (CR 608.2h), and the mill happens whether or not the
// spell could be countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ec9475f8-235a-4c85-9b96-9c2085a9b732",
		Name:           "Induce Paranoia",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"B"},
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets:        effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			id := targets[0].ID
			var controller uuid.UUID
			if s := ctx.Game.StackItemForEffect(id); s != nil {
				controller = s.Controller
			}
			mv, _ := ctx.Game.LastKnownSpellManaValueForEffect(id)
			black := ctx.ManaSpentOfColor("B") > 0
			if err := (effects.CounterTarget{StackID: id}).Apply(ctx); err != nil {
				return err
			}
			if !black || controller == uuid.Nil || mv <= 0 {
				return nil
			}
			return effects.MillCards{Player: controller, N: mv}.Apply(ctx)
		},
	})
}
