package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escalate_register_test.go — checkEscalate refuses the declarations
// that would compile and then charge something the card does not
// print (#2126).

func TestCheckEscalateRefusesShapesTheCastPlanCannotPay(t *testing.T) {
	modes := func(esc *game.EscalateCost) *game.ModeSpec {
		return Escalating(ChooseOneOrMore(Mode("A."), Mode("B.")), esc)
	}
	for _, tc := range []struct {
		name string
		spec Spec
	}{
		{"one mode only", Spec{Name: "X", Modes: Escalating(ChooseOne(Mode("A."), Mode("B.")), EscalateMana("{1}"))}},
		{"empty cost", Spec{Name: "X", Modes: modes(&game.EscalateCost{Label: "Escalate"})}},
		{"no label", Spec{Name: "X", Modes: modes(&game.EscalateCost{ManaCost: "{1}"})}},
		{"negative discard", Spec{Name: "X", Modes: modes(&game.EscalateCost{DiscardCards: -1, Label: "Escalate"})}},
		{"unparseable mana", Spec{Name: "X", Modes: modes(&game.EscalateCost{ManaCost: "{zz}", Label: "Escalate"})}},
		{"taps in the mandatory slot", Spec{Name: "X", AdditionalCost: &game.AdditionalCost{TapCreatures: 1}}},
		{"taps in an optional cost", Spec{Name: "X", OptionalCosts: []game.AdditionalCost{{Optional: true, Key: "kicker", TapCreatures: 1}}}},
		{"taps beside teamwork", Spec{Name: "X", Modes: modes(EscalateTapCreature()), OptionalCosts: []game.AdditionalCost{Teamwork(2)}}},
		{"escalate on a trigger", Spec{Name: "X", Triggered: []game.TriggeredAbility{{Modes: modes(EscalateMana("{1}"))}}}},
		{"escalate on an activated ability", Spec{Name: "X", Activated: []ActivatedAbility{{Modes: modes(EscalateMana("{1}"))}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("checkEscalate accepted a malformed escalate declaration")
				}
			}()
			checkEscalate(tc.spec)
		})
	}

	for _, ok := range []*game.EscalateCost{EscalateMana("{G}"), EscalateDiscard(1), EscalateDiscard(2), EscalateTapCreature()} {
		checkEscalate(Spec{Name: "ok", Modes: modes(ok)})
	}
}
