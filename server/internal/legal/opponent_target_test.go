package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// opponent_target_test.go — #2055. A clause "of an opponent's choice"
// is not the controller's to pick: the enumerator names no target for
// it, offers the announcement only while some opponent could choose
// one, and the opponent's pick_target is an ordinary owed choice. The
// view marks the clause so the client's walk skips it the same way.

const oracleArena = "12326d23-1fb5-4084-a23a-d6e2c90ea36c"

func TestArenaEnumeratesOnlyTheControllersOwnTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	arena := battlefieldCard(g, active, game.Card{Name: "Arena", TypeLine: "Land", OracleID: oracleArena})
	mine := battlefieldCard(g, active, creature("Mine", "{G}", 3, 3))
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)

	// No opponent controls a creature: no opponent could choose, so the
	// ability is not offered at all (CR 601.2c).
	if got := activationsOf(legal.EnumerateFor(g, active.ID), arena); len(got) != 0 {
		t.Fatalf("Arena offered with no creature an opponent controls: %v", labels(got))
	}

	theirs := battlefieldCard(g, opp, creature("Theirs", "{G}", 2, 2))
	moves := activationsOf(legal.EnumerateFor(g, active.ID), arena)
	if len(moves) == 0 {
		t.Fatal("Arena is not offered once an opponent controls a creature")
	}
	for _, m := range moves {
		ids := boundTargetIDs(t, m)
		if len(ids) != 1 || ids[0] != mine.String() {
			t.Fatalf("%q names %v, want only the controller's own creature", m.Label, ids)
		}
	}
	dispatchAll(t, g, active.ID, moves)

	// The view marks the opponent's clause.
	v := protocol.ViewOfGameFor(g, active.ID.String())
	marked := false
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != arena.String() {
			continue
		}
		for _, ab := range c.ActivatedAbilities {
			for i, cl := range ab.Clauses {
				if cl.ChosenByOpponent != (i == 1) {
					t.Fatalf("clause %d chosen_by_opponent = %v", i, cl.ChosenByOpponent)
				}
				marked = marked || cl.ChosenByOpponent
			}
		}
	}
	if !marked {
		t.Fatal("the view does not mark the opponent's clause")
	}

	// Announce it; only one opponent controls a creature, so that
	// opponent owes the pick, and the enumerator offers it to them.
	a := actions.Action{Type: actions.Type(moves[0].Type), Player: active.ID, Caller: active.ID, Params: moves[0].Params}
	if err := actions.Dispatch(g, a); err != nil {
		t.Fatalf("dispatch %s: %v", moves[0].Label, err)
	}
	var picks []legal.Move
	for _, m := range legal.EnumerateFor(g, opp.ID) {
		if m.Type == string(actions.TypeResolveChoice) {
			picks = append(picks, m)
		}
	}
	if len(picks) == 0 {
		t.Fatal("the opponent is offered no answer to their pick")
	}
	for _, m := range picks {
		var p struct {
			Target *struct {
				ID string `json:"id"`
			} `json:"target"`
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params %s: %v", string(m.Params), err)
		}
		var named []string
		if p.Target != nil {
			named = append(named, p.Target.ID)
		}
		for _, x := range p.Targets {
			named = append(named, x.ID)
		}
		for _, id := range named {
			if id != theirs.String() {
				t.Fatalf("the opponent was offered %s, not a creature they control", id)
			}
		}
	}
	dispatchAll(t, g, opp.ID, picks)
	vo := protocol.ViewOfGameFor(g, opp.ID.String())
	for _, c := range vo.PendingChoices {
		if c.Kind == string(game.PendingChoicePickTarget) && (c.PickTarget == nil || c.PickTarget.ChosenFor != active.ID.String()) {
			t.Fatalf("the opponent's pick does not say whose ability it is for: %+v", c.PickTarget)
		}
	}
	_ = uuid.Nil
}
