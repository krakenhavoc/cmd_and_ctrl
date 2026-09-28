package game

import (
	"testing"

	"github.com/google/uuid"
)

// combat_lki_test.go — #1661, CR 603.10a: a leaving permanent's combat
// state rides its EventLTB, so a THIRD PARTY's "whenever an attacking
// creature dies" (Kardur, Doomscourge) can read it after the exit has
// cleared Card.AttackingTarget and Card.BlockingTarget.
//
// The facts are taken in battlefieldExitLocked and stamped at each of
// the three EventLTB emit sites, so every route is pinned here: the
// destroy / sacrifice route (executeBattlefieldLeaveLocked), the
// general effect route (executeZoneRouteLocked — a bounce, and a
// sandbox move to any zone but the battlefield or the stack), and the
// sandbox move's own exit (moveCardByRefLocked — a drag to the stack).

// ltbSince returns the EventLTB for cardID emitted after seq.
func ltbSince(t *testing.T, g *Game, seq uint64, cardID uuid.UUID) Event {
	t.Helper()
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventLTB && ev.CardID == cardID {
			return ev
		}
	}
	t.Fatalf("no EventLTB for %s after seq %d", cardID, seq)
	return Event{}
}

// combatExitRoutes are the three battlefield exits that emit EventLTB.
var combatExitRoutes = []struct {
	name string
	exit func(t *testing.T, g *Game, id uuid.UUID)
}{
	{"destroy", func(t *testing.T, g *Game, id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatalf("DestroyPermanentForEffect: %v", err)
			}
		})
	}},
	{"sacrifice", func(t *testing.T, g *Game, id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.SacrificePermanentForEffect(id); err != nil {
				t.Fatalf("SacrificePermanentForEffect: %v", err)
			}
		})
	}},
	{"bounce", func(t *testing.T, g *Game, id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.BounceToHandForEffect(id); err != nil {
				t.Fatalf("BounceToHandForEffect: %v", err)
			}
		})
	}},
	{"sandbox move", func(t *testing.T, g *Game, id uuid.UUID) {
		c, ok := g.battlefieldCardLocked(id)
		if !ok {
			t.Fatalf("setup: %s is not on the battlefield", id)
		}
		if err := g.MoveCardByID(
			ZoneRef{Kind: ZoneBattlefield},
			ZoneRef{Kind: ZoneGraveyard, Owner: c.Owner}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}},
	// The sandbox move's own exit — a battlefield-to-stack drag is the
	// one sandbox move that does not hand off to the zone route.
	{"sandbox move to stack", func(t *testing.T, g *Game, id uuid.UUID) {
		if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneStack}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}},
}

// A blocked attacker leaving by any route reports what it was
// attacking and that it was blocked; its blocker reports the attacker
// it was blocking. Neither card still says so — that is the gap.
func TestLTBCarriesCombatStateOnEveryExitRoute(t *testing.T) {
	for _, route := range combatExitRoutes {
		t.Run(route.name+"/attacker", func(t *testing.T) {
			g := newActiveGame(t)
			defender := g.Seats[1].ID
			attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
			blocker := pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
			blockAfterLockIn(t, g, attacker, blocker)

			seq := lastSeq(g)
			route.exit(t, g, attacker)
			ev := ltbSince(t, g, seq, attacker)
			if ev.AttackingTarget != defender {
				t.Errorf("AttackingTarget = %s, want the defending player %s", ev.AttackingTarget, defender)
			}
			if !ev.Blocked {
				t.Error("Blocked = false, want true — the attacker was blocked (CR 509.1h)")
			}
			if ev.BlockingTarget != uuid.Nil {
				t.Errorf("BlockingTarget = %s on an attacker, want none", ev.BlockingTarget)
			}
			if c, ok := g.LookupCardForEffect(attacker); ok && c.AttackingTarget != uuid.Nil {
				t.Error("the card still says it is attacking; the event is only needed because it does not")
			}
		})
		t.Run(route.name+"/blocker", func(t *testing.T) {
			g := newActiveGame(t)
			attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
			blocker := pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
			blockAfterLockIn(t, g, attacker, blocker)

			seq := lastSeq(g)
			route.exit(t, g, blocker)
			ev := ltbSince(t, g, seq, blocker)
			if ev.BlockingTarget != attacker {
				t.Errorf("BlockingTarget = %s, want the attacker %s", ev.BlockingTarget, attacker)
			}
			if ev.AttackingTarget != uuid.Nil || ev.Blocked {
				t.Errorf("a blocker reported attacking=%s blocked=%v, want neither", ev.AttackingTarget, ev.Blocked)
			}
		})
	}
}

// An unblocked attacker is attacking and not blocked; a creature that
// never entered combat reports nothing at all.
func TestLTBCombatStateForUnblockedAttackerAndBystander(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	bystander := pushCombatant(t, g, g.Seats[0], "Bystander", 1, 1)
	declareAttacks(t, g, attacker)

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(attacker); err != nil {
			t.Fatalf("destroy attacker: %v", err)
		}
		if err := g.DestroyPermanentForEffect(bystander); err != nil {
			t.Fatalf("destroy bystander: %v", err)
		}
	})
	if ev := ltbSince(t, g, seq, attacker); ev.AttackingTarget != g.Seats[1].ID || ev.Blocked {
		t.Errorf("unblocked attacker: attacking=%s blocked=%v, want seat 1 and false", ev.AttackingTarget, ev.Blocked)
	}
	if ev := ltbSince(t, g, seq, bystander); ev.AttackingTarget != uuid.Nil || ev.BlockingTarget != uuid.Nil || ev.Blocked {
		t.Errorf("bystander carried combat state %+v", ev)
	}
}

// CR 506.4: a creature removed from combat — here by a control change —
// is no longer attacking, so when it later dies its event says so.
func TestLTBCombatStateAbsentAfterRemovalFromCombat(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	declareAttacks(t, g, attacker)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), attacker, g.Seats[1].ID, g.UntilEndOfTurnDuration(), "test — steal") {
			t.Fatal("GainControlForEffect refused")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c, _ := g.battlefieldCardLocked(attacker); c.AttackingTarget != uuid.Nil {
		t.Fatal("setup: the control change did not remove the attacker from combat")
	}

	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(attacker); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if ev := ltbSince(t, g, seq, attacker); ev.AttackingTarget != uuid.Nil || ev.Blocked {
		t.Errorf("removed-from-combat creature reported attacking=%s blocked=%v, want neither", ev.AttackingTarget, ev.Blocked)
	}
}

// The facts live on the event and nowhere else, so the event log is
// what carries them through a persisted snapshot and an undo clone.
func TestLTBCombatStateSurvivesSnapshotAndClone(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Grizzly Bears", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	blockAfterLockIn(t, g, attacker, blocker)
	seq := lastSeq(g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(attacker); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	want := ltbSince(t, g, seq, attacker)

	_, restored := roundTrip(t, g)
	if got := ltbSince(t, restored, seq, attacker); got.AttackingTarget != want.AttackingTarget || got.Blocked != want.Blocked {
		t.Errorf("snapshot round trip: attacking=%s blocked=%v, want %s %v",
			got.AttackingTarget, got.Blocked, want.AttackingTarget, want.Blocked)
	}
	clone := g.Clone()
	if got := ltbSince(t, clone, seq, attacker); got.AttackingTarget != want.AttackingTarget || got.Blocked != want.Blocked {
		t.Errorf("clone: attacking=%s blocked=%v, want %s %v",
			got.AttackingTarget, got.Blocked, want.AttackingTarget, want.Blocked)
	}
}
