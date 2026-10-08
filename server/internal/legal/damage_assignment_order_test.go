package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// damage_assignment_order_test.go — #2692. The one damage assignment
// the enumerator offers kills what the attacker's power can buy, not
// whatever lies down the declared blocker order: CR 510.1c lets the
// attacking player divide the damage among the blockers as they
// choose.

type assignmentAnswer struct {
	Assignments []struct {
		BlockerID string `json:"blocker_id"`
		Amount    int    `json:"amount"`
	} `json:"assignments"`
	TrampleTo int `json:"trample_to_player"`
}

// blockedCombat seats an attacker for the active player, blocks it
// with the defender's creatures in the order given, and stops at the
// combat damage step's assignment prompt.
func blockedCombat(t *testing.T, power, tough int, kw []string, blockers [][2]int) (*game.Game, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	atk := battlefieldCard(g, me, game.Card{
		Name: "Mary", TypeLine: "Creature — Pirate", Power: power, Toughness: tough, Keywords: kw,
	})
	ids := make([]uuid.UUID, len(blockers))
	for i, pt := range blockers {
		ids[i] = battlefieldCard(g, def, game.Card{
			Name: "Blocker", TypeLine: "Creature — Soldier", Power: pt[0], Toughness: pt[1],
		})
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, id := range ids {
		if err := g.DeclareBlocker(id, atk); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	return g, me.ID, ids
}

// assignmentMove is the one damage-assignment move offered to seat.
func assignmentMove(t *testing.T, g *game.Game, seat uuid.UUID) (legal.Move, map[uuid.UUID]int, int) {
	t.Helper()
	var found []legal.Move
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeResolveChoice {
			found = append(found, m)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d damage-assignment moves offered, want 1", len(found))
	}
	var a assignmentAnswer
	if err := json.Unmarshal(found[0].Params, &a); err != nil {
		t.Fatal(err)
	}
	got := map[uuid.UUID]int{}
	for _, e := range a.Assignments {
		got[uuid.MustParse(e.BlockerID)] = e.Amount
	}
	return found[0], got, a.TrampleTo
}

func onBattlefield(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestDamageAssignmentKillsWhatThePowerCanBuy(t *testing.T) {
	// Review game 2's seq 325: a 3/3 blocked by a 2/4, then a 1/1. Down
	// the declared order the 2/4 takes all 3 and nothing dies.
	g, me, ids := blockedCombat(t, 3, 3, nil, [][2]int{{2, 4}, {1, 1}})
	big, small := ids[0], ids[1]
	m, got, trample := assignmentMove(t, g, me)
	if got[small] != 1 || got[big] != 2 || trample != 0 {
		t.Fatalf("offered %v (trample %d), want 1 to the 1/1 and 2 to the 2/4", got, trample)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(m.Type), Player: m.Player, Caller: me, Params: m.Params,
	}); err != nil {
		t.Fatalf("the engine refused the offered assignment %s: %v", m.Params, err)
	}
	if c := onBattlefield(g, small); c != nil {
		t.Errorf("the 1/1 survived with %d damage", c.DamageMarked)
	}
	if c := onBattlefield(g, big); c == nil || c.DamageMarked != 2 {
		t.Errorf("the 2/4 = %+v, want alive with 2 damage", c)
	}
}

func TestDamageAssignmentPrefersTheMoreValuableKill(t *testing.T) {
	// 4 power over a 1/1, a 1/1 and a 3/3: the 3/3 and a 1/1 (worth
	// 7 + 3) beat the two 1/1s (3 + 3).
	g, me, ids := blockedCombat(t, 4, 4, nil, [][2]int{{1, 1}, {1, 1}, {3, 3}})
	m, got, _ := assignmentMove(t, g, me)
	if got[ids[2]] != 3 || got[ids[0]] != 1 || got[ids[1]] != 0 {
		t.Fatalf("offered %v, want 3 to the 3/3 and 1 to the first 1/1", got)
	}
	dispatchAll(t, g, me, []legal.Move{m})
}

func TestDamageAssignmentTramplesOnlyPastLethal(t *testing.T) {
	g, me, ids := blockedCombat(t, 6, 6, []string{"trample"}, [][2]int{{2, 4}, {1, 1}})
	m, got, trample := assignmentMove(t, g, me)
	if got[ids[0]] != 4 || got[ids[1]] != 1 || trample != 1 {
		t.Fatalf("offered %v (trample %d), want lethal to both and 1 over", got, trample)
	}
	dispatchAll(t, g, me, []legal.Move{m})

	// Short of lethal to both, the trampler kills the one worth more
	// and nothing goes over (CR 702.19b).
	g, me, ids = blockedCombat(t, 4, 4, []string{"trample"}, [][2]int{{1, 1}, {2, 4}})
	m, got, trample = assignmentMove(t, g, me)
	if got[ids[1]] != 4 || got[ids[0]] != 0 || trample != 0 {
		t.Fatalf("offered %v (trample %d), want all 4 on the 2/4", got, trample)
	}
	dispatchAll(t, g, me, []legal.Move{m})
}
