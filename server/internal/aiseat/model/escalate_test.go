package model

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// escalate_test.go walks ADR 0033 §5's five triggers one at a time.
// Each case starts from a window that escalates for NO reason, so a
// firing trigger is the thing the case added and nothing else.

func has(reasons []string, want string) bool {
	for _, r := range reasons {
		if r == want {
			return true
		}
	}
	return false
}

// quietPolicy has every threshold pushed out of reach, so the only
// triggers that can fire are the ones a test arms.
func quietPolicy() *Policy {
	cfg := DefaultConfig()
	cfg.Fallback = &stubB{}
	cfg.Log = testLogger()
	cfg.ThreatPower, cfg.ThreatCreature, cfg.DangerLife = 999, 999, 0
	cfg.Epsilon = 0
	return New(cfg)
}

func TestAQuietWindowFiresNothing(t *testing.T) {
	p := quietPolicy()
	if got := p.escalationReasons(castWindow(), nil); len(got) != 0 {
		t.Errorf("a routine window escalated for %v", got)
	}
}

func TestStackTargetingThisSeatEscalates(t *testing.T) {
	p := quietPolicy()
	in := castWindow()
	in.View.StackItems = []protocol.StackItemView{{
		ID: "s1", Kind: "spell", Controller: oppSeat.String(), Label: "Murder",
		Targets: []protocol.TargetRefView{{Kind: "player", ID: meSeat.String()}},
	}}
	if got := p.escalationReasons(in, nil); !has(got, ReasonStackTargetsMe) {
		t.Errorf("reasons = %v, want %s", got, ReasonStackTargetsMe)
	}

	// Aimed at one of this seat's permanents, which is the same
	// decision wearing a different hat.
	in = castWindow()
	in.View.Battlefield.Cards = []protocol.CardView{
		{InstanceID: "mine", Name: "Bear", Controller: meSeat.String(), TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	}
	in.View.StackItems = []protocol.StackItemView{{
		ID: "s1", Controller: oppSeat.String(), Label: "Murder",
		Targets: []protocol.TargetRefView{{Kind: "card", ID: "mine"}},
	}}
	if got := p.escalationReasons(in, nil); !has(got, ReasonStackTargetsMe) {
		t.Errorf("reasons = %v, want %s", got, ReasonStackTargetsMe)
	}
}

// The bot's OWN spell on the stack, pointed at its own permanent, is
// not an attack on it.
func TestOwnStackItemDoesNotEscalate(t *testing.T) {
	p := quietPolicy()
	in := castWindow()
	in.View.Battlefield.Cards = []protocol.CardView{
		{InstanceID: "mine", Name: "Bear", Controller: meSeat.String(), TypeLine: "Creature — Bear"},
	}
	in.View.StackItems = []protocol.StackItemView{{
		ID: "s1", Controller: meSeat.String(), Label: "Giant Growth",
		Targets: []protocol.TargetRefView{{Kind: "card", ID: "mine"}},
	}}
	if got := p.escalationReasons(in, nil); has(got, ReasonStackTargetsMe) {
		t.Errorf("the bot's own pump spell escalated: %v", got)
	}
}

func TestCombatEscalates(t *testing.T) {
	p := quietPolicy()
	for _, kind := range []legal.Kind{legal.KindAttack, legal.KindBlock} {
		in := castWindow()
		in.Moves = append(in.Moves, mv(kind, "a combat declaration", ""))
		if got := p.escalationReasons(in, nil); !has(got, ReasonCombat) {
			t.Errorf("%s: reasons = %v, want %s", kind, got, ReasonCombat)
		}
	}
}

func TestRemovalAgainstAThreateningBoardEscalates(t *testing.T) {
	in := castWindow()
	in.View.Battlefield.Cards = []protocol.CardView{
		{InstanceID: "big", Name: "Wurm", Controller: oppSeat.String(), TypeLine: "Creature — Wurm", Power: 7, Toughness: 7},
	}
	in.Moves = append(in.Moves,
		mv(legal.KindCast, "Cast Murder targeting Wurm", `{"instance_id":"murder","targets":[{"kind":"card","id":"big"}]}`))

	// Below the threshold: a 7-power board with the bar at 999 is not
	// a reason to spend a frontier call.
	if got := quietPolicy().escalationReasons(in, nil); has(got, ReasonRemovalVsThreat) {
		t.Errorf("escalated under the threshold: %v", got)
	}

	// At the shipped thresholds it fires.
	cfg := DefaultConfig()
	cfg.Fallback, cfg.Log, cfg.Epsilon = &stubB{}, testLogger(), 0
	if got := New(cfg).escalationReasons(in, nil); !has(got, ReasonRemovalVsThreat) {
		t.Errorf("reasons = %v, want %s", got, ReasonRemovalVsThreat)
	}
}

// Removal is only interesting when it is pointed at somebody else.
// A spell the bot may aim at its own permanent is a pump, and a
// frontier call for a pump is exactly the leak the funnel exists to
// stop.
func TestRemovalPointedAtMyOwnPermanentDoesNotEscalate(t *testing.T) {
	in := castWindow()
	in.View.Battlefield.Cards = []protocol.CardView{
		{InstanceID: "big", Name: "Wurm", Controller: oppSeat.String(), TypeLine: "Creature — Wurm", Power: 7, Toughness: 7},
		{InstanceID: "mine", Name: "Bear", Controller: meSeat.String(), TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	}
	in.Moves = append(in.Moves,
		mv(legal.KindCast, "Cast Giant Growth targeting Bear", `{"instance_id":"gg","targets":[{"kind":"card","id":"mine"}]}`))
	cfg := DefaultConfig()
	cfg.Fallback, cfg.Log, cfg.Epsilon = &stubB{}, testLogger(), 0
	if got := New(cfg).escalationReasons(in, nil); has(got, ReasonRemovalVsThreat) {
		t.Errorf("a pump on my own creature escalated: %v", got)
	}
}

func TestCloseCallEscalates(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Fallback, cfg.Log = &stubB{}, testLogger()
	cfg.ThreatPower, cfg.ThreatCreature, cfg.DangerLife = 999, 999, 0
	cfg.Epsilon = 0.5
	p := New(cfg)

	tight := []heuristic.Candidate{{Index: 1, Value: 4.0}, {Index: 2, Value: 3.7}}
	if got := p.escalationReasons(castWindow(), tight); !has(got, ReasonCloseCall) {
		t.Errorf("reasons = %v, want %s", got, ReasonCloseCall)
	}
	clear := []heuristic.Candidate{{Index: 1, Value: 8.0}, {Index: 2, Value: 1.0}}
	if got := p.escalationReasons(castWindow(), clear); has(got, ReasonCloseCall) {
		t.Errorf("a clear winner escalated: %v", got)
	}
	// Two moves the scorer thinks are both worthless is not a close
	// call, it is a pass. Escalating there would burn a frontier call
	// on doing nothing.
	worthless := []heuristic.Candidate{{Index: 1, Value: 0}, {Index: 2, Value: -1}}
	if got := p.escalationReasons(castWindow(), worthless); has(got, ReasonCloseCall) {
		t.Errorf("a tie between two bad moves escalated: %v", got)
	}
}

func TestModalXAndMultiTargetEscalate(t *testing.T) {
	p := quietPolicy()
	cases := map[string]string{
		"modal":         `{"instance_id":"c","modes":[1]}`,
		"x cost":        `{"instance_id":"c","x_value":4}`,
		"two targets":   `{"instance_id":"c","targets":[{"kind":"card","id":"a"},{"kind":"card","id":"b"}]}`,
		"one target":    `{"instance_id":"c","targets":[{"kind":"card","id":"a"}]}`,
		"no parameters": `{"instance_id":"c"}`,
	}
	want := map[string]bool{"modal": true, "x cost": true, "two targets": true}
	for name, params := range cases {
		in := castWindow()
		in.Moves = append(in.Moves, mv(legal.KindCast, "Cast something", params))
		got := has(p.escalationReasons(in, nil), ReasonComplexChoice)
		if got != want[name] {
			t.Errorf("%s: complex-choice fired = %v, want %v", name, got, want[name])
		}
	}
}

func TestWidePickTargetChoiceEscalates(t *testing.T) {
	p := quietPolicy()
	in := castWindow()
	in.Moves = []legal.Move{
		mv(legal.KindChoice, "Target A", `{"choice_id":"c1"}`),
		mv(legal.KindChoice, "Target B", `{"choice_id":"c1"}`),
	}
	in.View.PendingChoices = []protocol.PendingChoiceView{{
		ID: "c1", Kind: "pick_target", Chooser: meSeat.String(),
		PickTarget: &protocol.LegalTargetsView{Cards: []string{"a", "b", "c", "d", "e"}},
	}}
	if got := p.escalationReasons(in, nil); !has(got, ReasonComplexChoice) {
		t.Errorf("reasons = %v, want %s", got, ReasonComplexChoice)
	}

	// Four is not more than four.
	in.View.PendingChoices[0].PickTarget.Cards = []string{"a", "b", "c", "d"}
	if got := p.escalationReasons(in, nil); has(got, ReasonComplexChoice) {
		t.Errorf("a four-candidate choice escalated: %v", got)
	}

	// Somebody else's choice is not this seat's problem.
	in.View.PendingChoices[0].PickTarget.Cards = []string{"a", "b", "c", "d", "e"}
	in.View.PendingChoices[0].Chooser = oppSeat.String()
	if got := p.escalationReasons(in, nil); has(got, ReasonComplexChoice) {
		t.Errorf("another seat's choice escalated: %v", got)
	}
}

// The `strong` tier escalates everything Layer A did not settle.
func TestStrongTierEscalatesEverySurvivingWindow(t *testing.T) {
	cfg := StrongConfig()
	cfg.Fallback, cfg.Log = &stubB{}, testLogger()
	cfg.ThreatPower, cfg.ThreatCreature, cfg.DangerLife = 999, 999, 0
	cfg.Epsilon = 0
	p := New(cfg)
	if got := p.escalationReasons(castWindow(), nil); !has(got, ReasonTier) {
		t.Errorf("reasons = %v, want %s", got, ReasonTier)
	}
	if p.Name() != TierStrong {
		t.Errorf("name = %q", p.Name())
	}
}

// Reasons are sorted, because they are a map key in the aggregate
// counters and an unsorted list would split one trigger across two
// labels depending on iteration order.
func TestReasonsAreSorted(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Fallback, cfg.Log = &stubB{}, testLogger()
	p := New(cfg)
	in := castWindow()
	in.Moves = append(in.Moves,
		mv(legal.KindAttack, "Attack Bot1 with Bear", ""),
		mv(legal.KindCast, "Cast something modal", `{"instance_id":"c","modes":[1]}`))
	got := p.escalationReasons(in, nil)
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Fatalf("reasons are not sorted: %v", got)
		}
	}
}
