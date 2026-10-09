package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thrull Wizard — Creature — Thrull Wizard {2}{B}, 1/1:
//
//	"{1}{B}: Counter target black spell unless that spell's controller
//	 pays {B} or {3}."
//
// "Unless that spell's controller pays {B} or {3}" is one choice with
// three answers (CR 118.12a, #2854): pay {B}, pay {3}, or pay nothing
// and the spell is countered. It is an option_pick to the spell's
// controller whose two payments carry their mana cost; a payment they
// cannot make is not offered (CR 118.3), and the chosen one is paid
// through the auto-tapper. An option_pick holds the table while it is
// open, so the spell cannot resolve before the answer. If the spell has
// left the stack anyway, the counter does nothing.
//
// The ability can be activated again on the same spell, each one asking
// for its own payment (the 2004-10-04 ruling).
//
// The continuation is a registered key (OptionPickThen), so a table
// waiting on the answer is still a restore point.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3cae0e4b-1827-4ed2-832f-98f8e79941a5",
		Name:         "Thrull Wizard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Counter target black spell unless that spell's controller pays {B} or {3}.",
			Cost:    ManaCost("{1}{B}"),
			Targets: TargetSpell("target black spell", OfColor("B")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				t := targets[0]
				spell := g.StackItemForEffect(t.ID)
				if spell == nil {
					return nil
				}
				return PickOption{
					Player:   spell.Controller,
					Question: "Thrull Wizard — pay {B} or {3}, or your spell is countered?",
					Options: []game.ChoiceOption{
						{Label: "Pay nothing: the spell is countered"},
						{Label: "Pay {B}", ManaCost: "{B}"},
						{Label: "Pay {3}", ManaCost: "{3}"},
					},
					ThenKey: thrullWizardAnswer,
					Carry:   []uuid.UUID{t.ID},
				}.Apply(ctx)
			},
		}},
	})
}

// thrullWizardAnswer is the registered continuation. Its key is an
// on-disk identity: never renamed, never reused.
var thrullWizardAnswer = OptionPickThen("option-pick/thrull-wizard-pay", thrullWizardAnswered)

// thrullWizardAnswered counters Carry[0] unless a payment was made. A
// controller who left the game took the spell with them (CR 800.4a).
func thrullWizardAnswered(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Option == nil || r.Option.ManaCost != "" {
		return nil
	}
	stackID := r.Carry[0]
	if ctx.Game.StackItemForEffect(stackID) == nil {
		return nil
	}
	if err := ctx.Game.CounterTargetForEffect(stackID); err != nil && !errors.Is(err, game.ErrCardNotOnStack) {
		return err
	}
	return nil
}
