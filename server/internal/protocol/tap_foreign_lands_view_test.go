package protocol

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_foreign_lands_view_test.go — the view half of #6204 (ADR 0106 §1
// amendment 2026-11-25): under "you may tap lands you don't control for
// mana" (Piracy) the holder's copy of another player's land marks its {T}
// row `any_player` (the client's "the viewer may activate this row"),
// shows the "only to cast spells" tag its mana will carry, and the
// holder's digest lists it. The controller's and a bystander's copies are
// the public row.
func TestForeignLandRowIsTheHoldersToActivate(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	controller := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bystander := g.Seats[(g.Turn.ActiveSeat+2)%4]
	mountain := put(g.Battlefield, controller, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	knowTheTable(g)
	g.WithWriteLock(func() { g.AllowTapForeignLandsThisTurnForEffect(active.ID) })

	av := ViewOfGameFor(g, active.ID.String())
	row := manaRowOf(t, av, mountain)
	if !row.AnyPlayer {
		t.Error("the holder's copy of the Mountain's row is not marked any_player")
	}
	if !slices.Contains(row.Restrictions, game.ManaRestrictCast) {
		t.Errorf("the holder's row restrictions = %v, want %q", row.Restrictions, game.ManaRestrictCast)
	}
	if av.LegalActions == nil {
		t.Fatal("the active seat got no digest")
	}
	e := av.LegalActions.Sources[mountain.String()]
	if e == nil || !slices.Equal(e.ManaAbilities, []string{row.Ref}) {
		t.Fatalf("holder's digest entry = %+v, want mana_abilities [%s]", e, row.Ref)
	}
	for _, p := range []*game.Player{controller, bystander} {
		r := manaRowOf(t, ViewOfGameFor(g, p.ID.String()), mountain)
		if r.AnyPlayer || len(r.Restrictions) != 0 {
			t.Errorf("%s's copy = %+v, want the public row", p.Name, r)
		}
	}
}
