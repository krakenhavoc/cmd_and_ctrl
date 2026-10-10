package rules_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// value_only_test.go is ADR 0142 decision 6's Layer A rule: a window
// with something on the stack whose only alternatives to passing are
// rows declared `value` is a pass.

const (
	clueID  = "00000000-0000-0000-0000-0000000000c1"
	spellID = "00000000-0000-0000-0000-0000000000f1"
)

func clueView(answers string, stackTargets ...string) protocol.GameView {
	a := []string{answers}
	clue := protocol.CardView{
		InstanceID: clueID, Name: "Clue", Controller: seat.String(), Owner: seat.String(), TypeLine: "Token Artifact — Clue",
		ActivatedAbilities: []protocol.ActivatedAbilityView{{Index: 0, Label: "{2}, Sacrifice this artifact: Draw a card.",
			ManaCost: "{2}", SacrificeSelf: true, Purpose: &protocol.PurposeView{Draws: 1, Answers: &a}}},
	}
	other := "22222222-2222-2222-2222-222222222222"
	div := protocol.CardView{InstanceID: spellID, Name: "Divination", Controller: other, Owner: other, TypeLine: "Sorcery"}
	item := protocol.StackItemView{ID: "i", Kind: "spell", Controller: other, SourceCardID: spellID}
	for _, id := range stackTargets {
		item.Targets = append(item.Targets, protocol.TargetRefView{Kind: "card", ID: id})
	}
	return protocol.GameView{
		Battlefield: protocol.ZoneView{Kind: "battlefield", Cards: []protocol.CardView{clue}},
		Stack:       protocol.ZoneView{Kind: "stack", Cards: []protocol.CardView{div}},
		StackItems:  []protocol.StackItemView{item},
	}
}

func crackClue(t *testing.T) legal.Move {
	b, err := json.Marshal(map[string]any{"source_card_id": clueID, "ability_index": 0})
	if err != nil {
		t.Fatal(err)
	}
	return legal.Move{Type: legal.TypeActivateAbility, Kind: legal.KindActivate, Player: seat, Label: "Crack the Clue", Params: b}
}

func TestValueOnlyPassesAValueRowWithSomethingOnTheStack(t *testing.T) {
	in := aiseat.Input{Seat: seat, View: clueView("value"), Moves: []legal.Move{pass(), crackClue(t), mv(legal.KindMana, "Tap Mountain for R")}}
	v := rules.Resolve(in)
	if !v.Absorbed() || v.Rule != rules.RuleValueOnly || v.Index != 0 {
		t.Fatalf("got %+v, want a value-only pass", v)
	}
	// Layer A never answers differently from the heuristic it fronts.
	d, err := heuristic.New().Decide(context.Background(), in)
	if err != nil || d.Index != v.Index {
		t.Errorf("the heuristic chose %d (%s, %v), Layer A %d", d.Index, d.Reason, err, v.Index)
	}
}

func TestValueOnlyEscalates(t *testing.T) {
	cases := map[string]aiseat.Input{
		"nothing on the stack": {Seat: seat, View: func() protocol.GameView {
			v := clueView("value")
			v.StackItems, v.Stack = nil, protocol.ZoneView{}
			return v
		}(), Moves: []legal.Move{pass(), crackClue(t)}},
		"a row that answers":      {Seat: seat, View: clueView("protect"), Moves: []legal.Move{pass(), crackClue(t)}},
		"the source is a target":  {Seat: seat, View: clueView("value", clueID), Moves: []legal.Move{pass(), crackClue(t)}},
		"a cast is on offer too":  {Seat: seat, View: clueView("value"), Moves: []legal.Move{pass(), crackClue(t), mv(legal.KindCast, "Cast Opt")}},
		"the move is interacting": {Seat: seat, View: clueView("value"), Moves: []legal.Move{pass(), func() legal.Move { m := crackClue(t); m.Interacts = true; return m }()}},
	}
	for name, in := range cases {
		if v := rules.Resolve(in); v.Rule == rules.RuleValueOnly {
			t.Errorf("%s: value-only absorbed it", name)
		}
	}
}

// NoValueOnly turns the rule off, for a heuristic seated with
// PriceAnswers off.
func TestValueOnlyOffUnderNoValueOnly(t *testing.T) {
	in := aiseat.Input{Seat: seat, View: clueView("value"), Moves: []legal.Move{pass(), crackClue(t)}}
	if v := rules.ResolveWith(in, rules.Options{NoValueOnly: true}); v.Absorbed() {
		t.Errorf("got %+v, want an escalation", v)
	}
}
