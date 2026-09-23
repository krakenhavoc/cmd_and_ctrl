package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_total_lock_view_test.go — #1200's projection half.
//
// `life_total_locked` is public and unredacted for the reason
// `keywords` and `emblems` are: the lock changes what every player at
// the table may do, and a viewer who can see the Bolt move no life
// but not why is strictly worse off.

// TestLifeTotalLockIsProjectedToEveryViewer — both halves of the
// rule, since the two stores are read by one predicate: a stored
// grant (Teferi's Protection) and a derived one (Platinum Emperion)
// both reach the wire, and a seat with neither carries nothing.
func TestLifeTotalLockIsProjectedToEveryViewer(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	g.WithWriteLock(func() {
		g.GrantLifeTotalLockForEffect(me.ID, "Teferi's Protection — your life total can't change",
			uuid.Nil, g.UntilYourNextTurnDuration(me.ID))
	})

	for _, viewer := range []*game.Player{me, opp} {
		view := ViewOfGameFor(g, viewer.ID.String())
		for _, seat := range view.Seats {
			switch seat.ID {
			case me.ID.String():
				if !seat.LifeTotalLocked {
					t.Errorf("viewer %s sees no lock on the locked seat", viewer.Name)
				}
			case opp.ID.String():
				if seat.LifeTotalLocked {
					t.Errorf("viewer %s sees a lock on a seat that has none", viewer.Name)
				}
			}
		}
	}
}

// TestDerivedLifeTotalLockReachesTheWire — the other store. It is not
// written to the engine's Player at all, so the field has to be
// EFFECTIVE, computed on every projection the way max_hand_size is.
func TestDerivedLifeTotalLockReachesTheWire(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]

	const emperion = "test-view-life-lock-emperion"
	prev := game.CatalogPlayerLifeTotalLocked
	game.CatalogPlayerLifeTotalLocked = func(id string) bool { return id == emperion }
	t.Cleanup(func() { game.CatalogPlayerLifeTotalLocked = prev })

	c := game.NewCard("Platinum Emperion", me.ID)
	c.TypeLine = "Artifact Creature — Golem"
	c.OracleID = emperion
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.Battlefield.PushTop(c)

	view := ViewOfGameFor(g, me.ID.String())
	for _, seat := range view.Seats {
		if seat.ID == me.ID.String() && !seat.LifeTotalLocked {
			t.Error("the Emperion's controller carries no lock on the wire")
		}
	}
}

// TestUnlockedSeatOmitsTheFieldEntirely — `omitempty` on a bool, so
// nearly every seat in nearly every game pays nothing for it.
func TestUnlockedSeatOmitsTheFieldEntirely(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	view := ViewOfGameFor(g, me.ID.String())

	raw, err := json.Marshal(view.Seats[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, present := decoded["life_total_locked"]; present {
		t.Errorf("an unlocked seat put life_total_locked on the wire: %s", raw)
	}
}
