package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_slot_view_test.go — #4128: the wire half of
// game.TargetSpec.OpponentSlot. A clause bound to the n-th opponent
// ships min 1 while that seat holds a live opponent and min 0 once it
// does not, so the picker never demands a pick for a player who is not
// at the table. The bounds ride each clause, which is what the client's
// multi-clause picker reads.

const viewSlotOracle = "test-view-opponent-slot"

func slotSpec() *game.TargetSpec {
	clause := func(n int) *game.TargetSpec {
		return &game.TargetSpec{Mode: "permanent", Label: "target permanent",
			Zones: []game.ZoneKind{game.ZoneBattlefield}, OpponentSlot: n, Min: 1, Max: 1}
	}
	return clause(1).Then(clause(2), clause(3))
}

func slotMins(t *testing.T, g *game.Game) []int {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Slot Thing", TypeLine: "Sorcery", OracleID: viewSlotOracle,
			ManaCost: "{1}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})
	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.LegalTargets == nil {
		t.Fatal("the card carries no legal targets")
	}
	var mins []int
	for _, cl := range c.Clauses {
		mins = append(mins, cl.Min)
	}
	return mins
}

func TestOpponentSlotClausesAreRequiredOnlyForOpponentsAtTheTable(t *testing.T) {
	g := buildActiveGame(t) // two seats: one opponent
	withTargetSpec(t, viewSlotOracle, slotSpec)
	got := slotMins(t, g)
	want := []int{1, 0, 0}
	if len(got) != len(want) {
		t.Fatalf("clause mins = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("clause mins = %v, want %v (a two-player table has one opponent)", got, want)
			break
		}
	}
}

func TestOpponentSlotClauseStopsBeingRequiredWhenItsOpponentIsEliminated(t *testing.T) {
	g := buildActiveGame(t)
	withTargetSpec(t, viewSlotOracle, slotSpec)
	if got := slotMins(t, g); got[0] != 1 {
		t.Fatalf("with the opponent alive the first clause is required, got %v", got)
	}
	g.Seats[1].Eliminated = true
	if got := slotMins(t, g); got[0] != 0 {
		t.Errorf("with the opponent eliminated the first clause is not required, got %v", got)
	}
}
