package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// enlist_test.go — #5658, the enumerator half of enlist (CR 702.154):
// an attacker with enlist is offered each attack once more for every
// creature it may enlist, and the dispatcher accepts every one (#544).

type enlistParams struct {
	Target string `json:"target"`
	Enlist string `json:"enlist"`
}

// enlistMovesOf is the creature's attack moves that enlist, as
// "target enlisted" keys.
func enlistMovesOf(t *testing.T, moves []legal.Move, id uuid.UUID) map[string]legal.Move {
	t.Helper()
	out := map[string]legal.Move{}
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != id {
			continue
		}
		var p enlistParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.Enlist != "" {
			out[p.Target+" "+p.Enlist] = m
		}
	}
	return out
}

func TestEnlistCreatureIsOfferedATwinAttackPerCandidate(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	enlister := creature("Barkweave Crusher", "{3}{G}", 2, 5)
	enlister.Keywords = []string{game.KeywordEnlist}
	crusher := battlefieldCard(g, active, enlister)
	bear := battlefieldCard(g, active, creature("Bear", "{G}", 2, 2))
	elk := battlefieldCard(g, active, creature("Elk", "{2}{G}", 3, 3))
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, active.ID)
	plain, _ := attackMovesOf(t, moves, crusher)
	twins := enlistMovesOf(t, moves, crusher)
	if len(plain) == 0 || len(twins) != 2*len(plain) {
		t.Fatalf("%d plain attacks and %d enlist twins, want two twins per target", len(plain), len(twins))
	}
	for key, m := range twins {
		if !strings.Contains(m.Label, "and enlist Bear (tap it to add its power)") && !strings.Contains(m.Label, "and enlist Elk (tap it to add its power)") {
			t.Errorf("enlist move %s label %q", key, m.Label)
		}
		if m.AlwaysLegal {
			t.Error("an enlist move is never the owed answer (CR 508.1d)")
		}
	}
	if n := len(enlistMovesOf(t, moves, bear)); n != 0 {
		t.Errorf("a creature without enlist is offered %d enlist moves", n)
	}
	dispatchAll(t, g, active.ID, moves)

	// Bear attacks: it is no longer a candidate, and Elk still is.
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.DeclareAttacker(bear, def.ID); err != nil {
		t.Fatal(err)
	}
	moves = legal.EnumerateFor(g, active.ID)
	for key := range enlistMovesOf(t, moves, crusher) {
		if strings.HasSuffix(key, bear.String()) {
			t.Error("an attacking creature is offered as the enlisted creature")
		}
		if !strings.HasSuffix(key, elk.String()) {
			t.Errorf("unexpected enlist move %s", key)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}
