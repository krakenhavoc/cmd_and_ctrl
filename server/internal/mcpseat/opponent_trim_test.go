package mcpseat

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// overBudgetTable is bigBoard (three padded opponents, far past the compact
// budget) plus one extra opponent the caller fills.
func overBudgetTable(me string) (*protocol.GameView, string) {
	v := bigBoard(me)
	id := uuid.NewString()
	v.Seats = append(v.Seats, protocol.PlayerView{ID: id, Name: "Token Player", Seat: 4, Life: 30})
	return v, id
}

func battlefieldLineOf(board, seatName string) string {
	in := false
	for _, l := range strings.Split(board, "\n") {
		if strings.Contains(l, seatName) && !strings.HasPrefix(l, "  ") {
			in = true
			continue
		}
		if in && strings.HasPrefix(strings.TrimSpace(l), "battlefield:") {
			return l
		}
		if in && !strings.HasPrefix(l, "  ") {
			in = false
		}
	}
	return ""
}

// #3123: the opponent pre-trim ran before boardtext grouped anything, so a
// token-heavy board lost permanents the grouped line would have fitted.
func TestOverBudgetOpponentKeepsEveryGroupedPermanent(t *testing.T) {
	me := uuid.NewString()
	v, opp := overBudgetTable(me)
	names := []string{"Avenger of Zendikar", "Sol Ring", "Doubling Season", "Temple of Mystery", "Lotus Cobra"}
	for i := 0; i < 22; i++ {
		v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{InstanceID: uuid.NewString(), Controller: opp,
			Name: "Plant", TypeLine: "Token Creature — Plant", Toughness: 1, IsToken: true})
	}
	for _, n := range names {
		v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{InstanceID: uuid.NewString(), Controller: opp,
			Name: n, TypeLine: "Artifact"})
	}
	board := compactBoard(v, me, sinceLog{}, nil)
	if len(board) > budgetCompact {
		t.Fatalf("compact board is %d bytes, budget %d", len(board), budgetCompact)
	}
	line := battlefieldLineOf(board, "Token Player")
	if !strings.Contains(line, "Plant 0/1 x22") {
		t.Errorf("the 22 Plants are not one grouped entry:\n%s", board)
	}
	for _, n := range names {
		if !strings.Contains(line, n) {
			t.Errorf("%s missing from the opponent's line:\n%s", n, line)
		}
	}
	if strings.Contains(line, "more") {
		t.Errorf("a within-cap opponent got a tail:\n%s", line)
	}
}

func TestOverBudgetOpponentEndsWithBoardtextsExactTail(t *testing.T) {
	me := uuid.NewString()
	v, opp := overBudgetTable(me)
	for i := 0; i < 30; i++ {
		v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{InstanceID: uuid.NewString(), Controller: opp,
			Name: fmt.Sprintf("Distinct Artifact %d", i), TypeLine: "Artifact"})
	}
	board := compactBoard(v, me, sinceLog{}, nil)
	line := battlefieldLineOf(board, "Token Player")
	want := "… and 18 more (call get_state(detail: \"full\"))"
	if !strings.HasSuffix(line, want) {
		t.Errorf("opponent line does not end with %q:\n%s", want, line)
	}
	if strings.Contains(line, "and 18 more,") {
		t.Errorf("a fake card stands in for the tail:\n%s", line)
	}
}
