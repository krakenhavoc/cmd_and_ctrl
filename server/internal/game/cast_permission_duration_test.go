package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_permission_duration_test.go — #945: a granted cast permission
// carries ADR 0063's game.Duration and nothing else, so the boundary
// it ends on is the boundary a continuous effect with the same clause
// would end on.
//
// One test per kind a permission can carry, each driven through the
// REAL rotation (PassTurn goes through rotation.go's seam, so the
// cleanup sweep and the turn-begin sweep both run) rather than by
// poking Turn.Number — the round counter is exactly the thing the
// swap stopped being the answer.

// exiledWithGrant drops a fresh instant into exile carrying `perm`
// and returns its instance ID. The grant goes through the one write
// path, so a zero Duration is stamped "until end of this turn".
func exiledWithGrant(t *testing.T, g *Game, owner *Player, name string, perm CastPermission) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{R}"
	id := c.InstanceID
	g.WithWriteLock(func() {
		g.Exile.PushTop(c)
		g.markCardKnownInZoneLocked(g.Exile, id)
		if !g.GrantCastPermissionOverCardForEffect(id, perm) {
			t.Fatalf("GrantCastPermissionOverCardForEffect(%s) found no card", name)
		}
	})
	return id
}

// castableFromExileBy asks the ONE predicate the cast path, the view
// and the bot enumerator all ask.
func castableFromExileBy(g *Game, id, player uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() {
		c, found := g.cardInZoneLocked(g.Exile, id)
		if !found {
			return
		}
		ok = g.CastPermissionForLocked(player, c, ZoneExile) != nil
	})
	return ok
}

// "Until end of turn" ends at the cleanup step of the turn it was
// made in (CR 514.2) — the kind impulse exile, Snapcaster, cascade
// and a Siege's face grant all carry.
func TestPermissionUntilEndOfTurnEndsAtThatTurnsCleanup(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := exiledWithGrant(t, g, me, "Impulsed Bolt", CastPermission{Player: me.ID})

	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("the grant is live on the turn it was made")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if castableFromExileBy(g, id, me.ID) {
		t.Error("an until-end-of-turn grant outlived its turn")
	}
	if n := len(me.CastPermissions); n != 0 {
		t.Errorf("the cleanup sweep left %d permissions behind", n)
	}
}

// "Until the end of your next turn" (Reckless Impulse, Wrenn's
// Resolve, Prosper) is the clause the batch 19 / 20 delayed-trigger
// re-stamp existed to fake. Four seats is what makes it a real test:
// the grant has to ride out three opponents' turns and then the whole
// of the holder's own.
func TestPermissionUntilEndOfYourNextTurnRidesOutThreeOpponents(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var window Duration
	g.WithWriteLock(func() { window = g.UntilEndOfYourNextTurnDuration(me.ID) })
	id := exiledWithGrant(t, g, me, "Reckless Hit", CastPermission{Player: me.ID, Duration: window})

	for i, who := range []string{"my own turn", "opponent 1", "opponent 2", "opponent 3"} {
		if !castableFromExileBy(g, id, me.ID) {
			t.Fatalf("the grant died during %s", who)
		}
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn %d: %v", i, err)
		}
	}
	// Back on my turn — the one the clause names the END of.
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatalf("setup: expected to be back on my own turn, active seat is %d", g.Turn.ActiveSeat)
	}
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("the grant is live through the whole of your next turn")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if castableFromExileBy(g, id, me.ID) {
		t.Error("the grant outlived the end of your next turn")
	}
}

// "Until your next turn" is the shorter sibling: it ends as that turn
// BEGINS (CR 500.1), which is the boundary onTurnBeganLocked sweeps.
// Nothing prints it on a cast permission yet; the model accepts it,
// and this is what "accepts" has to mean.
func TestPermissionUntilYourNextTurnEndsAsThatTurnBegins(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var window Duration
	g.WithWriteLock(func() { window = g.UntilYourNextTurnDuration(me.ID) })
	id := exiledWithGrant(t, g, me, "Borrowed Bolt", CastPermission{Player: me.ID, Duration: window})

	for i := 0; i < 3; i++ {
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn %d: %v", i, err)
		}
		if !castableFromExileBy(g, id, me.ID) {
			t.Fatalf("the grant died on opponent %d's turn", i+1)
		}
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if castableFromExileBy(g, id, me.ID) {
		t.Error("an until-your-next-turn grant survived the turn it names")
	}
}

// Warp and foretell print a FLOOR, not a duration: "you may cast it
// from exile on a later turn" (CR 702.185a, CR 702.143a). Duration
// has no vocabulary for when an effect STARTS, which is why
// NotBeforeTurn survived the swap — and it composes with the
// zone-bound window rather than replacing it.
func TestPermissionNotBeforeTurnIsAFloorUnderAZoneBoundWindow(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := exiledWithGrant(t, g, me, "Warped Bear", CastPermission{
		Player:        me.ID,
		Duration:      WhileInZoneDuration(),
		NotBeforeTurn: g.Turn.Number + 1,
	})

	if castableFromExileBy(g, id, me.ID) {
		t.Error("a foretold / warped card is castable on the turn it was exiled")
	}
	// Round the table once: Turn.Number goes up as the rotation wraps
	// past seat 0, which is the counter CR 702.185a's floor is read
	// against.
	for i := 0; i < len(g.Seats); i++ {
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn %d: %v", i, err)
		}
	}
	if !castableFromExileBy(g, id, me.ID) {
		t.Errorf("the floor never lifted: turn %d, grant %+v",
			g.Turn.Number, exilePlayOf(g, id))
	}
}

// The zone-bound kind ends on a ZONE change rather than on a turn,
// and CR 400.7 is what says so: the permission names {instance,
// epoch}, so a card that leaves exile and comes back is a new object
// the grant no longer names. Nothing in durationExpiredLocked is
// involved, which is the point of the kind's comment.
func TestZoneBoundPermissionEndsOnTheZoneChange(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := exiledWithGrant(t, g, me, "Airbent Bear", CastPermission{
		Player: me.ID, Duration: WhileInZoneDuration(),
	})

	// Turns do not touch it.
	for i := 0; i < len(g.Seats)+1; i++ {
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn %d: %v", i, err)
		}
	}
	if !castableFromExileBy(g, id, me.ID) {
		t.Fatal("a zone-bound grant lapsed on a turn boundary")
	}

	// Out of exile and back: a new object (CR 400.7).
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneExile}, ZoneRef{Kind: ZoneHand, Owner: me.ID}, id); err != nil {
		t.Fatalf("exile -> hand: %v", err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, id); err != nil {
		t.Fatalf("hand -> exile: %v", err)
	}
	if castableFromExileBy(g, id, me.ID) {
		t.Error("the grant followed the card back into exile")
	}
	// And the husk is swept, so the slice does not grow for the rest
	// of the game.
	g.WithWriteLock(func() { g.sweepCastPermissionsLocked(false) })
	if n := len(me.CastPermissions); n != 0 {
		t.Errorf("the sweep left %d spent permissions behind", n)
	}
}

// Undo and the persisted snapshot both carry the Duration verbatim —
// it is pure data, which is what keeps Player.CastPermissions
// classified `carried` by snapshot_drift_test.go.
func TestPermissionDurationSurvivesCloneAndSnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zoneBound := exiledWithGrant(t, g, me, "Airbent Bear", CastPermission{
		Player: me.ID, Duration: WhileInZoneDuration(),
	})
	var window Duration
	g.WithWriteLock(func() { window = g.UntilEndOfYourNextTurnDuration(me.ID) })
	nextTurn := exiledWithGrant(t, g, me, "Reckless Hit", CastPermission{Player: me.ID, Duration: window})

	for _, tc := range []struct {
		name string
		make func(t *testing.T) *Game
	}{
		{"clone", func(t *testing.T) *Game { return g.Clone() }},
		{"snapshot round-trip", func(t *testing.T) *Game {
			restored, err := g.CaptureSnapshot().Restore()
			if err != nil {
				t.Fatalf("Restore: %v", err)
			}
			return restored
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.make(t)
			if got := exilePlayOf(out, zoneBound).Duration; got.Kind != WhileInZone {
				t.Errorf("zone-bound grant came back as %v", got.Kind)
			}
			got := exilePlayOf(out, nextTurn).Duration
			if got.Kind != UntilEndOfTurn || got.Player != me.ID ||
				got.ExpiresAfterTurnsBegun != window.ExpiresAfterTurnsBegun {
				t.Errorf("until-end-of-next-turn grant came back as %+v, want %+v", got, window)
			}
		})
	}
}
