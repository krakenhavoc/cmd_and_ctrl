package mcpseat

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attackView is the agent's own declare-attackers step, with a third
// seat so "all" has to pick one opponent.
func attackView(f *fakeServer, carol uuid.UUID) protocol.GameView {
	v := activeView(f)
	v.Seats = append(v.Seats, protocol.PlayerView{ID: carol.String(), Name: "Carol", Seat: 2, Life: 40})
	v.Turn.ActiveSeat, v.Turn.PriorityHolder, v.Turn.Step = 0, 0, stepDeclareAttackers
	return v
}

func attackMove(f *fakeServer, attacker uuid.UUID, target, label string, exert bool) legal.Move {
	p := map[string]any{"attacker": attacker.String(), "target": target}
	if exert {
		p["exert"] = true
	}
	return legal.Move{Type: legal.TypeDeclareAttacker, Player: f.playerID, Kind: legal.KindAttack, Label: label, Source: attacker, Params: mustJSON(p)}
}

func TestTheAttackWindowOffersAttackAllAndSaysWhatPassingDoes(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	joinFake(t, f, s)
	carol := uuid.New()
	v := attackView(f, carol)
	bob := f.oppID.String()
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	moves := []legal.Move{
		f.pass(),
		attackMove(f, a, bob, "Attack Bob with Bear", false),
		attackMove(f, a, carol.String(), "Attack Carol with Bear", false),
		attackMove(f, b, bob, "Attack Bob with Wolf", false),
		attackMove(f, b, bob, "Attack Bob with Wolf and exert it", true),
		attackMove(f, c, bob, "Attack Bob with Elk", false),
	}
	f.setState(v, moves, false)
	tok, text := decisionWindow(t, s)

	if !strings.Contains(text, "kind: attack") {
		t.Errorf("not an attack window:\n%s", text)
	}
	if !strings.Contains(text, "0: No attack: pass priority without declaring attackers") {
		t.Errorf("the pass does not say it declares no attack:\n%s", text)
	}
	// Bob: three creatures, one of which could be exerted. Carol: one
	// creature, which already has its own move.
	if !strings.Contains(text, "6: Attack «Bob» with all 3 creatures that can attack them (exerts none of them)") {
		t.Errorf("no attack-all at Bob:\n%s", text)
	}
	if strings.Count(text, "with all") != 1 {
		t.Errorf("an attack-all was offered for a single creature:\n%s", text)
	}

	r, _ := s.Act(context.Background(), ActInput{Window: tok, Move: 6})
	if !strings.Contains(r.Text, "status: accepted") {
		t.Fatalf("act: %s", r.Text)
	}
	sent := f.lastAction()
	if sent.Type != typeDeclareAttackers || sent.Player != f.playerID.String() {
		t.Fatalf("sent %+v", sent)
	}
	var p attackAllPayload
	if err := json.Unmarshal(sent.Params, &p); err != nil {
		t.Fatal(err)
	}
	want := []attackAllEntry{{a.String(), bob}, {b.String(), bob}, {c.String(), bob}}
	if len(p.Attackers) != len(want) || p.AutoTap {
		t.Fatalf("params %+v, want %+v without auto_tap", p, want)
	}
	for i := range want {
		if p.Attackers[i] != want[i] {
			t.Errorf("attacker %d = %+v, want %+v", i, p.Attackers[i], want[i])
		}
	}

	// Once a creature is attacking, the pass is the declaration's end.
	v.Battlefield.Cards = []protocol.CardView{{InstanceID: a.String(), Controller: f.playerID.String(), AttackingTarget: bob}}
	f.setState(v, []legal.Move{f.pass(), attackMove(f, c, bob, "Attack Bob with Elk", false)}, false)
	_, text = decisionWindow(t, s)
	if !strings.Contains(text, "0: Done declaring attackers: pass priority, attacking with 1 creature") {
		t.Errorf("the pass does not say it ends the declaration:\n%s", text)
	}
}

func TestAttackAllHonoursTheTaxTheLimitAndTheStep(t *testing.T) {
	me, bob := uuid.New(), uuid.New()
	v := &protocol.GameView{
		Seats: []protocol.PlayerView{{ID: me.String(), Name: "Agent", Seat: 0}, {ID: bob.String(), Name: "Bob", Seat: 1}},
		Turn:  protocol.TurnView{ActiveSeat: 0, Step: stepDeclareAttackers},
	}
	pass := legal.Move{Type: legal.TypePassPriority, Player: me, Kind: legal.KindPass, Label: "Pass priority", AlwaysLegal: true}
	attack := func(id uuid.UUID, extra map[string]any) legal.Move {
		p := map[string]any{"attacker": id.String(), "target": bob.String()}
		for k, x := range extra {
			p[k] = x
		}
		return legal.Move{Type: legal.TypeDeclareAttacker, Player: me, Kind: legal.KindAttack, Label: "Attack Bob", Source: id, Params: mustJSON(p)}
	}
	moves := []legal.Move{pass, attack(uuid.New(), nil), attack(uuid.New(), map[string]any{"auto_tap": true})}
	added := func(v *protocol.GameView, moves []legal.Move, partial bool) []legal.Move {
		out := withAttackDeclaration(v, me.String(), moves, partial)
		return out[len(moves):]
	}

	// Propaganda: the price is on the label, and the lands may pay it.
	v.Turn.AttackTargets = []protocol.AttackTargetView{{Kind: "player", ID: bob.String(), Tax: "{2}"}}
	got := added(v, moves, false)
	if len(got) != 1 || !strings.Contains(got[0].Label, "pays {2} each, {2}{2} in all") {
		t.Fatalf("taxed attack-all: %+v", got)
	}
	var p attackAllPayload
	if err := json.Unmarshal(got[0].Params, &p); err != nil || !p.AutoTap {
		t.Errorf("a taxed attack-all must let the lands pay: %s", got[0].Params)
	}
	// It does not grow on a second pass over the same window.
	if again := withAttackDeclaration(v, me.String(), append(moves, got...), false); len(again) != len(moves)+1 {
		t.Errorf("extended twice: %d moves", len(again))
	}

	// Silent Arbiter: two would break the limit, so no "all".
	one := 1
	v.Turn.AttackTargets = []protocol.AttackTargetView{{Kind: "player", ID: bob.String(), AttackLimit: &one}}
	if got := added(v, moves, false); len(got) != 0 {
		t.Errorf("attack-all over the limit: %+v", got)
	}
	v.Turn.AttackTargets = nil

	// A tax paid with life is not bundled.
	life := []legal.Move{pass, attack(uuid.New(), map[string]any{"phyrexian_life": 1}), attack(uuid.New(), nil)}
	if got := added(v, life, false); len(got) != 0 {
		t.Errorf("attack-all with a life payment: %+v", got)
	}

	// A list the server cut is not a list to say "all" from.
	if got := added(v, moves, true); len(got) != 0 {
		t.Errorf("attack-all on a partial list: %+v", got)
	}

	// Not the seat's own declaration: nothing changes.
	v.Turn.ActiveSeat = 1
	out := withAttackDeclaration(v, me.String(), moves, false)
	if len(out) != len(moves) || out[0].Label != "Pass priority" {
		t.Errorf("changed outside the seat's declaration: %+v", out)
	}
	v.Turn.ActiveSeat, v.Turn.Step = 0, "main1"
	if out := withAttackDeclaration(v, me.String(), moves, false); out[0].Label != "Pass priority" {
		t.Errorf("relabelled outside declare_attackers: %q", out[0].Label)
	}
}
