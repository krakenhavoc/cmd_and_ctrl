package rules_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// --- builders -------------------------------------------------------

var seat = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func mv(kind legal.Kind, label string) legal.Move {
	return legal.Move{Kind: kind, Label: label, Player: seat}
}

func pass() legal.Move { return mv(legal.KindPass, "Pass priority") }

func input(moves ...legal.Move) aiseat.Input {
	return aiseat.Input{Seat: seat, Moves: moves}
}

// --- the filter -----------------------------------------------------

func TestResolve(t *testing.T) {
	cases := []struct {
		name  string
		in    aiseat.Input
		want  rules.Outcome
		rule  string
		index int
	}{
		{
			name: "no moves is not a decision",
			in:   input(),
			want: rules.Escalate,
			rule: rules.RuleNoMoves,
		},
		{
			name:  "a lone pass is forced",
			in:    input(pass()),
			want:  rules.Take,
			rule:  rules.RuleForced,
			index: 0,
		},
		{
			name:  "one legal move of any kind is forced",
			in:    input(mv(legal.KindChoice, "Discard Mountain")),
			want:  rules.Take,
			rule:  rules.RuleForced,
			index: 0,
		},
		{
			name:  "floating mana is never the play",
			in:    input(pass(), mv(legal.KindMana, "Tap Mountain for R"), mv(legal.KindMana, "Tap Island for U")),
			want:  rules.Take,
			rule:  rules.RuleManaOnly,
			index: 0,
		},
		{
			name: "copies of one land are interchangeable",
			in: input(pass(),
				mv(legal.KindLand, "Play Mountain"),
				mv(legal.KindLand, "Play Mountain"),
				mv(legal.KindMana, "Tap Mountain for R")),
			want:  rules.Take,
			rule:  rules.RuleSameLand,
			index: 1,
		},
		{
			name: "two different lands is a colour decision",
			in: input(pass(),
				mv(legal.KindLand, "Play Mountain"),
				mv(legal.KindLand, "Play Island")),
			want: rules.Escalate,
			rule: rules.RuleNone,
		},
		{
			name: "anything castable escalates",
			in: input(pass(),
				mv(legal.KindMana, "Tap Mountain for R"),
				mv(legal.KindCast, "Cast Lightning Bolt targeting Kess")),
			want: rules.Escalate,
			rule: rules.RuleNone,
		},
		{
			name: "an activated ability escalates",
			in:   input(pass(), mv(legal.KindActivate, "Activate Skullclamp")),
			want: rules.Escalate,
			rule: rules.RuleNone,
		},
		{
			// A blocks window reaches the seat BEFORE priority does,
			// so there is no pass to take and nothing the rules
			// settle. This is the case that would be a bug if Layer A
			// answered it: "take the first block" is chump-blocking
			// with everything.
			name: "a blocks-only window is a decision",
			in: input(mv(legal.KindBlock, "Block Wurm with Bear"),
				mv(legal.KindBlock, "Block Wurm with Drake")),
			want: rules.Escalate,
			rule: rules.RuleNone,
		},
		{
			name: "a mulligan is a decision",
			in:   input(mv(legal.KindMulligan, "Keep"), mv(legal.KindMulligan, "Mulligan to 6")),
			want: rules.Escalate,
			rule: rules.RuleNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rules.Resolve(tc.in)
			if got.Outcome != tc.want {
				t.Fatalf("outcome = %v, want %v (rule %q)", got.Outcome, tc.want, got.Rule)
			}
			if got.Rule != tc.rule {
				t.Errorf("rule = %q, want %q", got.Rule, tc.rule)
			}
			if got.Absorbed() && got.Index != tc.index {
				t.Errorf("index = %d, want %d", got.Index, tc.index)
			}
			if got.Absorbed() && got.Reason == "" {
				t.Error("an absorbed window with no reason tells the log nothing")
			}
		})
	}
}

// Every index Layer A returns has to be a move. This is the one bug
// class that would reach the dispatcher as an out-of-range panic
// rather than as bad play.
func TestResolveNeverReturnsAnIndexThatIsNotAMove(t *testing.T) {
	windows := []aiseat.Input{
		input(pass()),
		input(pass(), mv(legal.KindMana, "Tap Mountain for R")),
		input(pass(), mv(legal.KindLand, "Play Mountain"), mv(legal.KindLand, "Play Mountain")),
		input(pass(), mv(legal.KindCast, "Cast Bear")),
		input(),
	}
	for i, in := range windows {
		v := rules.Resolve(in)
		if !v.Absorbed() {
			continue
		}
		if v.Index < 0 || v.Index >= len(in.Moves) {
			t.Errorf("window %d: absorbed with index %d against %d moves", i, v.Index, len(in.Moves))
		}
	}
}

// --- the meter ------------------------------------------------------

func TestMeterCountsEveryWindowAndRule(t *testing.T) {
	var m rules.Meter
	m.Observe(rules.Resolve(input(pass())))                                     // forced
	m.Observe(rules.Resolve(input(pass(), mv(legal.KindMana, "Tap Mountain")))) // mana-only
	m.Observe(rules.Resolve(input(pass(), mv(legal.KindCast, "Cast Bear"))))    // escalate

	st := m.Stats()
	if st.Windows != 3 {
		t.Fatalf("windows = %d, want 3", st.Windows)
	}
	if st.Absorbed != 2 {
		t.Fatalf("absorbed = %d, want 2", st.Absorbed)
	}
	if got := st.Rate(); got < 0.66 || got > 0.67 {
		t.Errorf("rate = %.3f, want 2/3", got)
	}
	if st.ByRule[rules.RuleForced] != 1 || st.ByRule[rules.RuleManaOnly] != 1 || st.ByRule[rules.RuleNone] != 1 {
		t.Errorf("byRule = %v", st.ByRule)
	}
	if rules := st.Rules(); len(rules) != 3 {
		t.Errorf("Rules() = %v, want three entries", rules)
	}
	m.Reset()
	if m.Stats().Windows != 0 {
		t.Error("Reset left counters behind")
	}
}

// A meter that has seen nothing has proved nothing: an empty rate
// must not read as a perfect one, or a broken harness would look like
// a passing exit criterion.
func TestEmptyMeterRateIsZeroNotOne(t *testing.T) {
	var m rules.Meter
	if got := m.Stats().Rate(); got != 0 {
		t.Errorf("rate of an empty meter = %v, want 0", got)
	}
}

// A nil Meter is the "instrumentation off" configuration and must not
// panic anywhere.
func TestNilMeterIsUsable(t *testing.T) {
	var m *rules.Meter
	m.Observe(rules.Resolve(input(pass())))
	if st := m.Stats(); st.Windows != 0 {
		t.Errorf("nil meter reported %d windows", st.Windows)
	}
	m.Reset()
}

// --- the decorator --------------------------------------------------

// countingPolicy is the inner policy: it records whether it was
// reached at all.
type countingPolicy struct {
	calls    int
	concedes bool
}

func (p *countingPolicy) Name() string { return "counting" }

func (p *countingPolicy) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	p.calls++
	return aiseat.Decision{Index: 0, Reason: "inner"}, nil
}

func (p *countingPolicy) ShouldConcede(aiseat.Input) bool { return p.concedes }

func TestFilterAnswersAbsorbedWindowsWithoutWakingTheInnerPolicy(t *testing.T) {
	inner := &countingPolicy{}
	var m rules.Meter
	f := rules.NewFilter(inner, &m)

	d, err := f.Decide(context.Background(), input(pass(), mv(legal.KindMana, "Tap Mountain")))
	if err != nil {
		t.Fatal(err)
	}
	if d.Index != 0 {
		t.Errorf("index = %d, want the pass at 0", d.Index)
	}
	if inner.calls != 0 {
		t.Errorf("the inner policy was called %d times on a window Layer A settles", inner.calls)
	}

	if _, err := f.Decide(context.Background(), input(pass(), mv(legal.KindCast, "Cast Bear"))); err != nil {
		t.Fatal(err)
	}
	if inner.calls != 1 {
		t.Errorf("the inner policy was called %d times on a real decision, want 1", inner.calls)
	}
	if st := m.Stats(); st.Windows != 2 || st.Absorbed != 1 {
		t.Errorf("meter = %+v", st)
	}
}

func TestFilterForwardsConcede(t *testing.T) {
	inner := &countingPolicy{concedes: true}
	f := rules.NewFilter(inner, nil)
	if !f.ShouldConcede(input(pass())) {
		t.Error("the filter swallowed a concede; Layer A holds no opinions about positions")
	}
	inner.concedes = false
	if f.ShouldConcede(input(pass())) {
		t.Error("the filter invented a concede")
	}
}

func TestFilterNameCarriesTheTier(t *testing.T) {
	f := rules.NewFilter(&countingPolicy{}, nil)
	if f.Name() != "rules+counting" {
		t.Errorf("Name() = %q", f.Name())
	}
	f.Tier = "heuristic"
	if f.Name() != "heuristic" {
		t.Errorf("Name() = %q after setting Tier", f.Name())
	}
}
