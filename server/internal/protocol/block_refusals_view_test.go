package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_refusals_view_test.go — #3898: the declare-blockers stamps the
// board's click-to-block reads. A creature a pending defender could
// declare lists the attackers it could not block alone, with the
// validator's reason and sentence; an attacker that takes more than one
// blocker says how many.
func TestBlockRefusalsAndMinBlockersAreStampedForThePendingDefender(t *testing.T) {
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%3]
	push := func(owner *game.Player, name string, keywords ...string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
			Keywords: keywords,
		})
		return id
	}
	flyer := push(active, "Sky Bear", "flying")
	brute := push(active, "Menace Bear", "menace")
	plain := push(active, "Plain Bear")
	wall := push(def, "Wall")
	other := push(def, "Other Wall")
	for _, a := range []uuid.UUID{flyer, brute, plain} {
		if err := g.DeclareAttacker(a, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("step %s, want declare_blockers", g.Turn.Step)
	}
	cards := func() map[string]CardView {
		out := map[string]CardView{}
		for _, c := range ViewOfGame(g).Battlefield.Cards {
			out[c.InstanceID] = c
		}
		return out
	}
	reasons := func(c CardView) map[string]string {
		out := map[string]string{}
		for _, r := range c.BlockRefusals {
			out[r.Attacker] = r.Reason
			if r.Message == "" {
				t.Errorf("refusal of %s has no sentence", r.Attacker)
			}
		}
		return out
	}

	v := cards()
	got := reasons(v[wall.String()])
	if got[flyer.String()] != string(game.BlockReasonFlying) {
		t.Errorf("Wall on the flyer: reason %q, want flying", got[flyer.String()])
	}
	if got[brute.String()] != string(game.BlockReasonTooFewBlockers) {
		t.Errorf("Wall alone on the menace attacker: reason %q, want too_few_blockers", got[brute.String()])
	}
	if r, ok := got[plain.String()]; ok {
		t.Errorf("Wall on the plain attacker is refused (%q); it is legal", r)
	}
	if n := v[brute.String()].MinBlockers; n != 2 {
		t.Errorf("menace attacker min_blockers = %d, want 2", n)
	}
	if n := v[plain.String()].MinBlockers; n != 0 {
		t.Errorf("plain attacker min_blockers = %d, want 0 (omitted)", n)
	}
	if len(v[flyer.String()].BlockRefusals) != 0 {
		t.Error("an attacker carries block refusals")
	}

	// Two blockers on the menace attacker: a pairing already held is
	// not listed, and a third creature joining it alone would be legal.
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: wall, Attacker: brute},
		{Blocker: other, Attacker: brute},
	}); err != nil {
		t.Fatal(err)
	}
	v = cards()
	if _, ok := reasons(v[wall.String()])[brute.String()]; ok {
		t.Error("the pairing the declaration holds is still listed as refused")
	}
	if r := reasons(v[wall.String()])[flyer.String()]; r != string(game.BlockReasonFlying) {
		t.Errorf("a declared blocker's re-point onto the flyer: reason %q, want flying", r)
	}

	// Once the defender is done, nothing is stamped.
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatal(err)
	}
	for id, c := range cards() {
		if len(c.BlockRefusals) > 0 || c.MinBlockers > 0 {
			t.Errorf("%s still stamped after the declaration finished", id)
		}
	}
}
