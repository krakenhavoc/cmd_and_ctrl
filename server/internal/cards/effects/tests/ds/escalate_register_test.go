package testsds

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escalate_register_test.go — checkEscalate refuses the declarations
// that would compile and then charge something the card does not
// print (#2126).

func TestCheckEscalateRefusesShapesTheCastPlanCannotPay(t *testing.T) {
	modes := func(esc *game.EscalateCost) *game.ModeSpec {
		return effects.Escalating(effects.ChooseOneOrMore(effects.Mode("A."), effects.Mode("B.")), esc)
	}
	for _, tc := range []struct {
		name string
		spec effects.Spec
	}{
		{"one mode only", effects.Spec{Name: "X", Modes: effects.Escalating(effects.ChooseOne(effects.Mode("A."), effects.Mode("B.")), effects.EscalateMana("{1}"))}},
		{"empty cost", effects.Spec{Name: "X", Modes: modes(&game.EscalateCost{Label: "Escalate"})}},
		{"no label", effects.Spec{Name: "X", Modes: modes(&game.EscalateCost{ManaCost: "{1}"})}},
		{"negative discard", effects.Spec{Name: "X", Modes: modes(&game.EscalateCost{DiscardCards: -1, Label: "Escalate"})}},
		{"unparseable mana", effects.Spec{Name: "X", Modes: modes(&game.EscalateCost{ManaCost: "{zz}", Label: "Escalate"})}},
		{"taps in the mandatory slot", effects.Spec{Name: "X", AdditionalCost: &game.AdditionalCost{TapCreatures: 1}}},
		{"taps in an optional cost", effects.Spec{Name: "X", OptionalCosts: []game.AdditionalCost{{Optional: true, Key: "kicker", TapCreatures: 1}}}},
		{"taps beside teamwork", effects.Spec{Name: "X", Modes: modes(effects.EscalateTapCreature()), OptionalCosts: []game.AdditionalCost{effects.Teamwork(2)}}},
		{"escalate on a trigger", effects.Spec{Name: "X", Triggered: []game.TriggeredAbility{{Modes: modes(effects.EscalateMana("{1}"))}}}},
		{"escalate on an activated ability", effects.Spec{Name: "X", Activated: []effects.ActivatedAbility{{Modes: modes(effects.EscalateMana("{1}"))}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("checkEscalate accepted a malformed escalate declaration")
				}
			}()
			effects.CheckEscalate(tc.spec)
		})
	}

	for _, ok := range []*game.EscalateCost{effects.EscalateMana("{G}"), effects.EscalateDiscard(1), effects.EscalateDiscard(2), effects.EscalateTapCreature()} {
		effects.CheckEscalate(effects.Spec{Name: "ok", Modes: modes(ok)})
	}
}
