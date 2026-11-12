package game

import "testing"

// #5659: a UnitColorsOnly reduction takes its colour and never generic
// mana, where a plain coloured Unit follows CR 118.7b and 118.7c.
func TestColorsOnlyReductionNeverTakesGeneric(t *testing.T) {
	unit, err := ParseCost("{U}")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ cost, want, plain string }{
		{"{2}{U}{U}", "{2}{U}", "{2}{U}"},
		{"{2}{U}", "{2}", "{2}"},
		{"{2}", "{2}", "{1}"},
		{"{1}{G}", "{1}{G}", "{G}"},
		{"{G/U}", "", ""}, // the empty string is {0}
		{"{U/P}{1}", "{1}", "{1}"},
	} {
		cost, err := ParseCost(c.cost)
		if err != nil {
			t.Fatal(err)
		}
		if got := reduceColorsOnly(cost, 1, unit).String(); got != c.want {
			t.Errorf("%s less {U}, colours only = %s, want %s", c.cost, got, c.want)
		}
		if got := reduceBy(cost, 1, &unit).String(); got != c.plain {
			t.Errorf("%s less {U} by CR 118.7 = %s, want %s", c.cost, got, c.plain)
		}
	}
	// Twice the unit takes two, and stops when the colour runs out.
	cost, _ := ParseCost("{3}{U}")
	if got := reduceColorsOnly(cost, 2, unit).String(); got != "{3}" {
		t.Errorf("{3}{U} less {U}{U}, colours only = %s, want {3}", got)
	}
}

// The flag is refused on anything but a reduction by coloured symbols.
func TestColorsOnlyUnitProblem(t *testing.T) {
	parse := func(s string) *ParsedCost {
		u, err := ParseCost(s)
		if err != nil {
			t.Fatal(err)
		}
		return &u
	}
	for name, c := range map[string]struct {
		m  CostModifier
		ok bool
	}{
		"a blue reduction":  {CostModifier{Kind: CostReduction, Unit: parse("{U}"), UnitColorsOnly: true}, true},
		"no unit":           {CostModifier{Kind: CostReduction, UnitColorsOnly: true}, false},
		"an increase":       {CostModifier{Kind: CostIncrease, Unit: parse("{U}"), UnitColorsOnly: true}, false},
		"generic in unit":   {CostModifier{Kind: CostReduction, Unit: parse("{1}{U}"), UnitColorsOnly: true}, false},
		"a generic unit":    {CostModifier{Kind: CostReduction, Unit: parse("{2}"), UnitColorsOnly: true}, false},
		"a colourless unit": {CostModifier{Kind: CostReduction, Unit: parse("{C}"), UnitColorsOnly: true}, false},
		"a snow unit":       {CostModifier{Kind: CostReduction, Unit: parse("{S}"), UnitColorsOnly: true}, false},
		"a plain blue unit": {CostModifier{Kind: CostReduction, Unit: parse("{U}")}, true},
		"a plain generic":   {CostModifier{Kind: CostReduction}, true},
	} {
		if got := c.m.UnitProblem() == ""; got != c.ok {
			t.Errorf("%s: accepted = %v, want %v (%q)", name, got, c.ok, c.m.UnitProblem())
		}
	}
}
