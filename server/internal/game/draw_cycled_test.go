package game

import (
	"testing"

	"github.com/google/uuid"
)

// draw_cycled_test.go — #6169: a draw made while a cycling ability resolves
// names the cycled card (ReplacementEvent.DrawCycled), so a replacement can
// read "if a cycling ability of another nonland card would cause you to draw
// a card" (Unpredictable Cyclone). CR 702.29a: cycling is "[Cost], Discard
// this card: Draw a card." CR 614.1a: "instead" is a replacement effect.

func TestResolvingCyclingCardNamesOnlyTheCyclersDraw(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cycled := uuid.New()
	cases := []struct {
		name   string
		item   *StackItem
		drawer uuid.UUID
		want   uuid.UUID
	}{
		{"the cycler draws", &StackItem{Kind: StackItemActivated, Cycling: true, Controller: me.ID, SourceCardID: cycled}, me.ID, cycled},
		{"another player draws", &StackItem{Kind: StackItemActivated, Cycling: true, Controller: me.ID, SourceCardID: cycled}, opp.ID, uuid.Nil},
		{"not a cycling ability", &StackItem{Kind: StackItemActivated, Controller: me.ID, SourceCardID: cycled}, me.ID, uuid.Nil},
		{"nothing resolving", nil, me.ID, uuid.Nil},
	}
	for _, c := range cases {
		g.mu.Lock()
		g.beginResolvingLocked(c.item)
		got := g.resolvingCyclingCardLocked(c.drawer)
		g.resolving = nil
		g.mu.Unlock()
		if got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
	}
}

// The cycled card rides a draw's event into the draw-instead body's done,
// and survives being parked to a restore point.
func TestDrawDoneCarriesTheCycledCard(t *testing.T) {
	g := newActiveGame(t)
	cycled := uuid.New()
	ev := &ReplacementEvent{Kind: RepEventDraw, Actor: g.Seats[0].ID, DrawPlayer: g.Seats[0].ID, DrawCount: 1, DrawCycled: cycled}
	done := DrawDone{frame: newParkedDraw(ev)}
	g.mu.Lock()
	defer g.mu.Unlock()
	if got := done.Cycled(g); got != cycled {
		t.Fatalf("inline done names %s, want %s", got, cycled)
	}
	id := done.Park(g)
	restored := restoreParkedDraws(snapshotParkedDraws(g.parkedDraws))
	if restored[id] == nil || restored[id].cycled != cycled {
		t.Fatalf("the parked frame lost the cycled card: %+v", restored[id])
	}
	if got := DrawDoneByID(id).Cycled(g); got != cycled {
		t.Errorf("parked done names %s, want %s", got, cycled)
	}
}
