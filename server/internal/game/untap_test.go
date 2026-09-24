package game

import (
	"testing"

	"github.com/google/uuid"
)

// untap_test.go — #74. The untap step used to clear Card.Tapped in a
// bare loop and announce nothing, so "whenever a permanent becomes
// untapped" was unwritable and the step named after the action was
// the one place the action was invisible.

// withCatalogUntapStepPermissions swaps the package-level hook for
// the duration of the test, in the shape withCatalogTriggers uses.
func withCatalogUntapStepPermissions(t *testing.T, fn func(oracleID string) []UntapStepPermission) {
	t.Helper()
	prev := CatalogUntapStepPermissions
	CatalogUntapStepPermissions = fn
	t.Cleanup(func() { CatalogUntapStepPermissions = prev })
}

// pushTappedPermanent parks a tapped permanent on the battlefield
// under `controller`, with an oracle ID a catalog hook can key on.
func pushTappedPermanent(g *Game, controller uuid.UUID, name, oracleID, typeLine string, tapped bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		OracleID:   oracleID,
		TypeLine:   typeLine,
		Owner:      controller,
		Controller: controller,
		Tapped:     tapped,
	})
	return id
}

// untapEventsFor counts EventUntapCard entries naming the card.
func untapEventsFor(g *Game, cardID uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventUntapCard && ev.CardID == cardID {
			n++
		}
	}
	return n
}

// The defect, pinned: the untap step's turn-based action announces
// every permanent it untaps.
func TestUntapStepAnnouncesEveryPermanentItUntaps(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	seat0 := g.Seats[0].ID
	tapped := pushTappedPermanent(g, seat0, "Tapped Rock", "", "Artifact", true)
	upright := pushTappedPermanent(g, seat0, "Upright Rock", "", "Artifact", false)

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })

	if n := untapEventsFor(g, tapped); n != 1 {
		t.Errorf("the untap step announces the permanent it untapped: %d events", n)
	}
	// CR 701.26b is a change of state: an upright permanent does not
	// become untapped, and Mesmeric Orb must not see one.
	if n := untapEventsFor(g, upright); n != 0 {
		t.Errorf("an already-untapped permanent does not become untapped: %d events", n)
	}
	if c, ok := battlefieldCardByID(g, tapped); !ok || c.Tapped {
		t.Error("it is untapped")
	}
}

// The event carries the permanent's controller, because "that
// permanent's controller mills a card" is the only player the event
// is ever about and the permanent may be gone by resolution.
func TestUntapEventCarriesThePermanentsController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	theirs := g.Seats[2].ID
	id := pushTappedPermanent(g, theirs, "Their Rock", "", "Artifact", true)

	if err := g.TapCard(id, false); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	for _, ev := range g.Events {
		if ev.Kind == EventUntapCard && ev.CardID == id {
			if ev.Actor != theirs {
				t.Errorf("Actor = %s, want the permanent's controller %s", ev.Actor, theirs)
			}
			return
		}
	}
	t.Fatal("no untap event")
}

// With no catalog wired — which is this package's normal state, and
// a server built without the effects blank import — the untap step
// untaps exactly what CR 502.3 says and nothing else.
func TestUntapStepWithoutCatalogUntapsOnlyTheActiveSeats(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	mine := pushTappedPermanent(g, g.Seats[0].ID, "Mine", "", "Artifact", true)
	theirs := pushTappedPermanent(g, g.Seats[1].ID, "Theirs", "", "Artifact", true)

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })

	if c, _ := battlefieldCardByID(g, mine); c.Tapped {
		t.Error("the active seat's permanent untaps")
	}
	if c, _ := battlefieldCardByID(g, theirs); !c.Tapped {
		t.Error("nobody else's does")
	}
}

// A declared UntapStepPermission widens CR 502.3's set, and the
// widening is scoped by BOTH predicates: AppliesTo picks the step,
// Untaps picks the permanents.
func TestUntapStepPermissionWidensTheSet(t *testing.T) {
	const oracle = "test-seedborn-probe"
	withCatalogUntapStepPermissions(t, func(id string) []UntapStepPermission {
		if id != oracle {
			return nil
		}
		return []UntapStepPermission{{
			Label: "probe: untap all permanents you control",
			AppliesTo: func(_ *Game, source *Card, activePlayer uuid.UUID) bool {
				return activePlayer != source.Controller
			},
			Untaps: func(_ *Game, source, target *Card) bool {
				return target.Controller == source.Controller
			},
		}}
	})

	g := newFourPlayerActiveGame(t)
	me, active := g.Seats[1].ID, 0
	pushTappedPermanent(g, me, "Probe", oracle, "Creature — Spirit", false)
	mine := pushTappedPermanent(g, me, "Mine", "", "Artifact", true)
	uninvolved := pushTappedPermanent(g, g.Seats[2].ID, "Third Seat", "", "Artifact", true)

	g.WithWriteLock(func() { g.performUntapStepLocked(active) })

	if c, _ := battlefieldCardByID(g, mine); c.Tapped {
		t.Error("the permission untaps its controller's permanents during another player's untap step")
	}
	if n := untapEventsFor(g, mine); n != 1 {
		t.Errorf("a permission untap announces like any other: %d events", n)
	}
	if c, _ := battlefieldCardByID(g, uninvolved); !c.Tapped {
		t.Error("a third seat's permanent is nobody's business")
	}

	// The same permission during its own controller's untap step
	// adds nothing — AppliesTo says "each OTHER player".
	retap := pushTappedPermanent(g, me, "Retap", "", "Artifact", true)
	g.WithWriteLock(func() { g.performUntapStepLocked(1) })
	if c, _ := battlefieldCardByID(g, retap); c.Tapped {
		t.Error("its controller's own untap step untaps it by CR 502.3 anyway")
	}
}

// An untap permission can untap a creature on an opponent's turn, but
// CR 302.6 still keys summoning sickness to its controller's turn-began
// boundary. The permission must not clear the marker as a side effect.
func TestOpponentUntapPermissionDoesNotClearSummoningSickness(t *testing.T) {
	const oracle = "test-seedborn-probe"
	withCatalogUntapStepPermissions(t, func(id string) []UntapStepPermission {
		if id != oracle {
			return nil
		}
		return []UntapStepPermission{{
			AppliesTo: func(_ *Game, source *Card, activePlayer uuid.UUID) bool {
				return activePlayer != source.Controller
			},
			Untaps: func(_ *Game, source, target *Card) bool {
				return target.Controller == source.Controller
			},
		}}
	})

	g := newFourPlayerActiveGame(t)
	me := g.Seats[1].ID
	pushTappedPermanent(g, me, "Probe", oracle, "Creature — Spirit", false)
	fresh := pushTappedPermanent(g, me, "Fresh Bear", "", "Creature — Bear", true)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fresh {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })

	c, ok := battlefieldCardByID(g, fresh)
	if !ok {
		t.Fatal("still on the battlefield")
	}
	if c.Tapped {
		t.Error("the permission untapped it")
	}
	if !HasSummoningSickness(&c) {
		t.Error("it is not this creature's controller's turn — it is still summoning-sick")
	}
}
