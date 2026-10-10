package legal_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// sacrifice_all_test.go — #2097: a cast whose additional cost is
// "sacrifice all creatures you control" (Soulblast) has ONE payment,
// the engine's own set, and every move names it — so a policy can price
// what it gives up — and says so in its label. Every move is one the
// engine accepts (#544).

const oracleSoulblast = "18d4c57b-e2bf-47a0-8823-c4a79498a7ff"

func soulblastMoves(t *testing.T, g *game.Game, seat *game.Player) (uuid.UUID, []legal.Move) {
	t.Helper()
	sb := handCard(seat, game.Card{Name: "Soulblast", TypeLine: "Instant", ManaCost: "{3}{R}{R}{R}", OracleID: oracleSoulblast})
	moves := legal.EnumerateFor(g, seat.ID)
	var mine []legal.Move
	for _, m := range moves {
		if m.Source == sb && m.Kind == legal.KindCast {
			mine = append(mine, m)
		}
	}
	return sb, mine
}

func TestSacrificeAllNamesEveryCreatureOnEveryMove(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 6, "Mountain")
	var want []string
	for i := 0; i < 3; i++ {
		want = append(want, battlefieldCard(g, seat, creature("Bear", "{2}", 2, 2)).String())
	}
	battlefieldCard(g, g.Seats[1], creature("Ogre", "{2}", 3, 3))
	sb, moves := soulblastMoves(t, g, seat)
	if len(moves) == 0 {
		t.Fatal("Soulblast is not offered")
	}
	slices.Sort(want)
	for _, p := range varSacMovesOf(t, moves, sb) {
		got := slices.Clone(p.SacrificeIDs)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Fatalf("sacrifice_ids = %v, want every creature the seat controls %v", p.SacrificeIDs, want)
		}
	}
	for _, m := range moves {
		if !strings.Contains(m.Label, "(Sacrifice all creatures you control: Bear, Bear, Bear)") {
			t.Fatalf("label %q does not say the cast sacrifices all three", m.Label)
		}
	}
	all := legal.EnumerateFor(g, seat.ID)
	dispatchAll(t, g, seat.ID, all)
}

// With no creatures the cost is paid with nothing (CR 118.3), so the
// cast is offered, naming none, and the label says so.
func TestSacrificeAllWithNoCreaturesIsStillOffered(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 6, "Mountain")
	sb, moves := soulblastMoves(t, g, seat)
	if len(moves) == 0 {
		t.Fatal("Soulblast with no creatures is not offered")
	}
	for _, p := range varSacMovesOf(t, moves, sb) {
		if len(p.SacrificeIDs) != 0 {
			t.Fatalf("sacrifice_ids = %v, want none", p.SacrificeIDs)
		}
	}
	if !strings.Contains(moves[0].Label, "(Sacrifice all creatures you control: none)") {
		t.Fatalf("label %q does not say nothing is sacrificed", moves[0].Label)
	}
	dispatchAll(t, g, seat.ID, legal.EnumerateFor(g, seat.ID))
}
