package game

import "testing"

// Table-driven coverage across every token shape the S15 parser
// should recognise. The table sits alongside the implementation
// because the cost parser is a closed spec — adding a new token type
// to cost.go should come with a new row here.

func TestParseCostTable(t *testing.T) {
	type want struct {
		generic   int
		required  []ColorRequirement
		xSlots    int
		phyrexian bool
		snow      bool
	}
	cases := []struct {
		in   string
		want want
	}{
		// Empty / whitespace-only — lands and cost-less cards.
		{"", want{}},
		{"  ", want{}},

		// Pure generic.
		{"{0}", want{generic: 0}},
		{"{3}", want{generic: 3}},
		{"{10}", want{generic: 10}},

		// Monocolored.
		{"{W}", want{required: []ColorRequirement{{Options: []string{"W"}}}}},
		{"{U}", want{required: []ColorRequirement{{Options: []string{"U"}}}}},
		{"{B}", want{required: []ColorRequirement{{Options: []string{"B"}}}}},
		{"{R}", want{required: []ColorRequirement{{Options: []string{"R"}}}}},
		{"{G}", want{required: []ColorRequirement{{Options: []string{"G"}}}}},
		{"{C}", want{required: []ColorRequirement{{Options: []string{"C"}}}}},

		// Lowercase input normalises.
		{"{r}", want{required: []ColorRequirement{{Options: []string{"R"}}}}},

		// Mixed generic + colored — classic Lightning Bolt / Bolt-alike costs.
		{"{R}", want{required: []ColorRequirement{{Options: []string{"R"}}}}},
		{"{1}{R}", want{generic: 1, required: []ColorRequirement{{Options: []string{"R"}}}}},
		{"{2}{W}{W}", want{generic: 2, required: []ColorRequirement{
			{Options: []string{"W"}}, {Options: []string{"W"}},
		}}},
		// Commander-level big spell.
		{"{2}{W}{U}{B}{R}{G}", want{generic: 2, required: []ColorRequirement{
			{Options: []string{"W"}}, {Options: []string{"U"}}, {Options: []string{"B"}},
			{Options: []string{"R"}}, {Options: []string{"G"}},
		}}},

		// X slots.
		{"{X}", want{xSlots: 1}},
		{"{X}{R}", want{xSlots: 1, required: []ColorRequirement{{Options: []string{"R"}}}}},
		{"{X}{X}", want{xSlots: 2}}, // Rosheen Meanderer et al.

		// Hybrid (two colors, pick either).
		{"{W/U}", want{required: []ColorRequirement{{Options: []string{"W", "U"}}}}},
		{"{B/R}", want{required: []ColorRequirement{{Options: []string{"B", "R"}}}}},

		// Phyrexian.
		{"{W/P}", want{
			phyrexian: true,
			required:  []ColorRequirement{{Options: []string{"W"}, Phyrexian: true}},
		}},
		{"{1}{W/P}{W/P}", want{
			generic:   1,
			phyrexian: true,
			required: []ColorRequirement{
				{Options: []string{"W"}, Phyrexian: true},
				{Options: []string{"W"}, Phyrexian: true},
			},
		}},

		// Hybrid Phyrexian — CR 107.4's ten symbols. One symbol, two
		// colour options, the life option (#787); the ten are walked
		// exhaustively in phyrexian_mana_test.go.
		{"{G/W/P}", want{
			phyrexian: true,
			required:  []ColorRequirement{{Options: []string{"G", "W"}, Phyrexian: true}},
		}},
		{"{1}{G}{G/W/P}{W}", want{
			generic:   1,
			phyrexian: true,
			required: []ColorRequirement{
				{Options: []string{"G"}},
				{Options: []string{"G", "W"}, Phyrexian: true},
				{Options: []string{"W"}},
			},
		}},

		// Two-mana hybrid — {2/W} (pay 2 generic or 1 white).
		{"{2/W}", want{required: []ColorRequirement{
			{Options: []string{"W"}, NumericAlt: 2, HasNumericAlt: true},
		}}},

		// Snow.
		{"{S}", want{snow: true, required: []ColorRequirement{{Options: []string{"C"}}}}},
		{"{1}{S}", want{
			generic:  1,
			snow:     true,
			required: []ColorRequirement{{Options: []string{"C"}}},
		}},

		// Whitespace between tokens (some deck exports add it).
		{"{1} {R}", want{generic: 1, required: []ColorRequirement{{Options: []string{"R"}}}}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseCost(tc.in)
			if err != nil {
				t.Fatalf("ParseCost(%q) unexpected err: %v", tc.in, err)
			}
			if got.Generic != tc.want.generic {
				t.Errorf("Generic: got %d, want %d", got.Generic, tc.want.generic)
			}
			if got.XSlots != tc.want.xSlots {
				t.Errorf("XSlots: got %d, want %d", got.XSlots, tc.want.xSlots)
			}
			if got.HasPhyrexian != tc.want.phyrexian {
				t.Errorf("HasPhyrexian: got %v, want %v", got.HasPhyrexian, tc.want.phyrexian)
			}
			if got.HasSnow != tc.want.snow {
				t.Errorf("HasSnow: got %v, want %v", got.HasSnow, tc.want.snow)
			}
			if len(got.Required) != len(tc.want.required) {
				t.Fatalf("Required len: got %d, want %d (%+v)", len(got.Required), len(tc.want.required), got.Required)
			}
			for i, r := range got.Required {
				w := tc.want.required[i]
				if len(r.Options) != len(w.Options) {
					t.Errorf("slot %d Options len: got %v, want %v", i, r.Options, w.Options)
					continue
				}
				for j := range r.Options {
					if r.Options[j] != w.Options[j] {
						t.Errorf("slot %d Options[%d]: got %q, want %q", i, j, r.Options[j], w.Options[j])
					}
				}
				if r.Phyrexian != w.Phyrexian {
					t.Errorf("slot %d Phyrexian: got %v, want %v", i, r.Phyrexian, w.Phyrexian)
				}
				if r.HasNumericAlt != w.HasNumericAlt {
					t.Errorf("slot %d HasNumericAlt: got %v, want %v", i, r.HasNumericAlt, w.HasNumericAlt)
				}
				if r.NumericAlt != w.NumericAlt {
					t.Errorf("slot %d NumericAlt: got %d, want %d", i, r.NumericAlt, w.NumericAlt)
				}
			}
		})
	}
}

// Invalid inputs should return an error rather than a silent zero.
func TestParseCostErrors(t *testing.T) {
	cases := []string{
		"{",       // unterminated brace
		"{R",      // unterminated brace with content
		"R}",      // missing opening brace
		"{}",      // empty token body
		"{foo}",   // multi-char nonsense
		"{W/X/U}", // a triple whose tail is not Phyrexian (CR 107.4 has no such symbol)
		"{W/U/Q}", // untap symbol in a hybrid tail
		"{1/W/P}", // no numeric-alt hybrid Phyrexian is printed
		"{1x}",    // mixed alphanumeric token body
	}
	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			if _, err := ParseCost(c); err == nil {
				t.Errorf("ParseCost(%q) unexpectedly succeeded", c)
			}
		})
	}
}

// TestParsedCostStringRoundTrips is #1190's renderer: the view stamps
// what AbilityManaCostForEffect / ManaAbilityManaCostForEffect
// actually charge by turning their ParsedCost back into a string, and
// this table is the inverse of TestParseCostTable's shapes — for
// every one of them, ParseCost(s).String() must be byte-identical to
// s, or the row the player sees would drift from the printed grammar
// the moment a modifier's arithmetic reproduced a shape ParseCost
// itself accepts.
//
// {0} is the one input the table does not round-trip: it parses to
// the same zero Generic an absent generic component does (mana_cost.go's
// own note on ParseCost), so there is nothing left after parsing to
// tell String the symbol was ever there. That loss is pre-existing
// and out of #1190's scope, which is the {S} one.
func TestParsedCostStringRoundTrips(t *testing.T) {
	cases := []string{
		"",
		"{3}",
		"{10}",
		"{W}",
		"{U}",
		"{B}",
		"{R}",
		"{G}",
		"{C}",
		"{1}{R}",
		"{2}{W}{W}",
		"{2}{W}{U}{B}{R}{G}",
		"{X}",
		"{X}{R}",
		"{X}{X}",
		"{W/U}",
		"{B/R}",
		"{W/P}",
		"{1}{W/P}{W/P}",
		"{G/W/P}",
		"{1}{G}{G/W/P}{W}",
		"{2/W}",
		// Snow (#1190): the whole point of the fix. {S} and {C} parse
		// to the same Options set (ColorRequirement{Options:{"C"}}),
		// so without a per-symbol Snow bit a mixed cost like this one
		// would print two "{C}"s and silently turn a snow-mana
		// ability into a plain colorless one.
		"{S}",
		"{1}{S}",
		"{1}{S}{C}",
		"{C}{S}",
	}
	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			parsed, err := ParseCost(s)
			if err != nil {
				t.Fatalf("ParseCost(%q): unexpected error: %v", s, err)
			}
			if got := parsed.String(); got != s {
				t.Errorf("ParseCost(%q).String() = %q, want %q (round trip broken)", s, got, s)
			}
		})
	}
}

// TestParsedCostStringZero pins the one documented non-round-trip and
// the empty-cost case String shares no braces with — a mana ability
// whose only cost component is {T} must render "", not "{0}" or a
// stray brace pair.
func TestParsedCostStringZero(t *testing.T) {
	if got := (ParsedCost{}).String(); got != "" {
		t.Errorf("the zero ParsedCost renders %q, want \"\"", got)
	}
	zero, err := ParseCost("{0}")
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", "{0}", err)
	}
	if got := zero.String(); got != "" {
		t.Errorf(`ParseCost("{0}").String() = %q, want "" — {0} and no generic component parse to the same cost`, got)
	}
}
