package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// registry_cost_modifier_test.go — the boot-time guards ADR 0048
// addendum §11 and §16 put on cost modifiers. Each shape below is a
// card-file mistake the engine would otherwise refuse at every cast;
// Register must refuse it at boot instead, and must not have put the
// spec in the registry when it does.

// unitOf parses a mana string into the Unit a modifier carries.
func unitOf(t *testing.T, s string) *game.ParsedCost {
	t.Helper()
	c, err := game.ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return &c
}

// registerPanics runs Register and reports the panic message, or "" if
// it did not panic.
func registerPanics(spec Spec) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg, _ = r.(string)
			if msg == "" {
				msg = "panicked with a non-string value"
			}
		}
	}()
	Register(spec)
	return ""
}

func TestRegisterRefusesMalformedCostModifiers(t *testing.T) {
	increaseWith := func(unit string) game.CostModifier {
		m := CostsMore(1, "surcharge")
		m.Unit = unitOf(t, unit)
		return m
	}
	cases := []struct {
		name string
		spec Spec
		want string
	}{
		{
			name: "a CostFloor in SelfCostModifiers",
			spec: Spec{SelfCostModifiers: []game.CostModifier{CostsAtLeast(3, "self floor")}},
			want: "CostFloor",
		},
		{
			name: "a Unit on a self CostReduction",
			spec: Spec{SelfCostModifiers: []game.CostModifier{func() game.CostModifier {
				m := CostsLess(1, "coloured reduction")
				m.Unit = unitOf(t, "{W}")
				return m
			}()}},
			want: "something other than an increase",
		},
		{
			name: "a Unit on a battlefield CostReduction",
			spec: Spec{CostModifiers: []game.CostModifier{func() game.CostModifier {
				m := CostsLess(1, "coloured reduction")
				m.Unit = unitOf(t, "{1}")
				return m
			}()}},
			want: "something other than an increase",
		},
		{
			name: "a Unit on a battlefield CostFloor",
			spec: Spec{CostModifiers: []game.CostModifier{func() game.CostModifier {
				m := CostsAtLeast(3, "coloured floor")
				m.Unit = unitOf(t, "{B}")
				return m
			}()}},
			want: "something other than an increase",
		},
		{
			name: "a Unit carrying {X}",
			spec: Spec{SelfCostModifiers: []game.CostModifier{increaseWith("{X}{R}")}},
			want: "{X}",
		},
		{
			name: "a Unit carrying a hybrid symbol",
			spec: Spec{SelfCostModifiers: []game.CostModifier{increaseWith("{1}{W/U}")}},
			want: "hybrid",
		},
		{
			name: "a Unit carrying a two-mana hybrid symbol",
			spec: Spec{CostModifiers: []game.CostModifier{increaseWith("{2/W}")}},
			want: "hybrid",
		},
		{
			name: "a Unit carrying a Phyrexian symbol",
			spec: Spec{SelfCostModifiers: []game.CostModifier{increaseWith("{W/P}")}},
			want: "Phyrexian",
		},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.spec.Name = tc.name
			tc.spec.OracleID = "test-registry-cost-modifier-guard-" + string(rune('a'+i))
			msg := registerPanics(tc.spec)
			if msg == "" {
				t.Fatalf("Register accepted %s", tc.name)
			}
			if !strings.Contains(msg, tc.want) {
				t.Errorf("panic %q does not mention %q", msg, tc.want)
			}
			if Has(tc.spec.OracleID) {
				t.Errorf("a refused spec was registered anyway")
			}
		})
	}
}

// The shapes the guards must let through: a strive unit of generic
// and single-colour symbols on an increase, a plain self reduction,
// and a floor in the battlefield slot.
func TestRegisterAcceptsWellFormedCostModifiers(t *testing.T) {
	spec := Spec{
		OracleID: "test-registry-cost-modifier-ok",
		Name:     "Well-formed modifiers",
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{1}{W}", "Strive"),
			CostsLess(1, "{1} less"),
		},
		CostModifiers: []game.CostModifier{CostsAtLeast(3, "Trinisphere")},
	}
	if msg := registerPanics(spec); msg != "" {
		t.Fatalf("Register refused a well-formed spec: %s", msg)
	}
}

// Every modifier the catalog actually registers passes the same check
// the engine applies at cast time, so no shipped card is one the engine
// would refuse.
func TestCatalogCostModifierUnitsAreApplicable(t *testing.T) {
	for _, s := range All() {
		for _, mods := range [][]game.CostModifier{s.CostModifiers, s.SelfCostModifiers} {
			for _, m := range mods {
				if why := m.UnitProblem(); why != "" {
					t.Errorf("%s: %q declares %s", s.Name, m.Label, why)
				}
			}
		}
	}
}
