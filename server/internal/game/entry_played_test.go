package game

import "testing"

// entry_played_test.go — #1326, CR 305.4: an ordinary land PLAY
// stamps Event.Played, and every "put" path leaves it unset. Both
// halves of the finisher are exercised here at the engine level —
// the catalog-level proof cards (Deep Gnome Terramancer, Horn of
// Greed, City of Traitors) live in internal/cards/effects.

// lastZoneMoveToBattlefield returns the newest EventZoneMove landing
// on the battlefield, or nil.
func lastZoneMoveToBattlefield(g *Game) *Event {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == EventZoneMove && ev.NewZone == ZoneBattlefield {
			return &g.Events[i]
		}
	}
	return nil
}

// An ordinary, UNPAUSED land play from hand — the common case, no
// shockland-style prompt in the way — stamps Event.Played. This is
// the branch that used to duplicate the shared finisher inline
// (mutations.go's CastSpell) rather than calling it, which would have
// left this exact case unmarked.
func TestLandPlayFromHandStampsEventPlayed(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]

	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatalf("dropLandFromHand: %v", err)
	}
	ev := lastZoneMoveToBattlefield(g)
	if ev == nil {
		t.Fatal("no EventZoneMove onto the battlefield was emitted")
	}
	if !ev.Played {
		t.Error("a land played from hand must stamp Event.Played")
	}
}

// A land an effect PUTS onto the battlefield from hand — never
// played, per CR 305.4 — leaves Event.Played unset.
func TestLandPutFromHandLeavesEventPlayedUnset(t *testing.T) {
	g := newWrapGame(t, 2)
	active := g.Seats[0]
	if len(active.Hand.Cards) == 0 {
		t.Fatal("hand is empty")
	}
	cardID := active.Hand.Cards[0].InstanceID

	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(cardID, HandEntryOptions{}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
	ev := lastZoneMoveToBattlefield(g)
	if ev == nil {
		t.Fatal("no EventZoneMove onto the battlefield was emitted")
	}
	if ev.Played {
		t.Error("a land PUT onto the battlefield by an effect must not stamp Event.Played")
	}
}

// A land token — created, not moved from any zone (CR 111.1) — was
// never played either, and announces through EventTokenCreated rather
// than EventZoneMove.
func TestLandTokenLeavesEventPlayedUnset(t *testing.T) {
	g := newWrapGame(t, 2)
	active := g.Seats[0]

	g.WithWriteLock(func() {
		tmpl := Card{Name: "Land Token Fixture", TypeLine: "Basic Land — Mountain"}
		if err := g.CreateTokenForEffect(active.ID, tmpl, 1); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	var found *Event
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == EventTokenCreated {
			found = &g.Events[i]
			break
		}
	}
	if found == nil {
		t.Fatal("no EventTokenCreated was emitted")
	}
	if found.Played {
		t.Error("a created token was never played")
	}
}
