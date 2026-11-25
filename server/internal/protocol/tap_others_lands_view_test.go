package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_others_lands_view_test.go — the view and digest half of #6204
// (Piracy): an opponent's land reaches the Piracy player with its mana row
// marked `tap_across` and listed in their digest; nobody else's copy is
// marked.
func TestPiracyMarksAnOpponentsLandRowForThePiracyPlayer(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%4]
	mountain := put(g.Battlefield, controller, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	knowTheTable(g)

	if row := manaRowOf(t, ViewOfGameFor(g, active.ID.String()), mountain); row.TapAcross {
		t.Fatal("without Piracy the row is marked tap_across")
	}
	g.WithWriteLock(func() { g.TapLandsYouDontControlThisTurnForEffect(uuid.Nil, active.ID, "Piracy") })
	av := ViewOfGameFor(g, active.ID.String())
	row := manaRowOf(t, av, mountain)
	if !row.TapAcross {
		t.Fatal("the Piracy player's copy of the opponent's Mountain is not marked tap_across")
	}
	if av.LegalActions == nil {
		t.Fatal("the active seat got no digest")
	}
	if e := av.LegalActions.Sources[mountain.String()]; e == nil || !slices.Equal(e.ManaAbilities, []string{row.Ref}) {
		t.Fatalf("the Piracy player's digest entry = %+v, want mana_abilities [%s]", e, row.Ref)
	}
	for _, p := range []*game.Player{controller, bystander} {
		if manaRowOf(t, ViewOfGameFor(g, p.ID.String()), mountain).TapAcross {
			t.Errorf("%s's copy is marked tap_across", p.Name)
		}
	}
}
