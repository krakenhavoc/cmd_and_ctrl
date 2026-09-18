package game

import (
	"testing"

	"github.com/google/uuid"
)

// token_existence_test.go — #596 / CR 704.5d. A token that leaves the
// battlefield for any other zone ceases to exist. The reported case
// was a bot's Cyclonic Rift bouncing a Bird token, which handed its
// controller a real card in hand.

// pushToken puts a token permanent onto the battlefield under
// `owner`'s control. The "Token" in the type line is the whole of
// token identity (Card.IsToken), exactly as the catalog's templates
// stamp it.
func pushToken(g *Game, owner *Player, name, typeLine string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.Power = 2
	c.Toughness = 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// settle runs the state-based check loop the way a priority boundary
// does. Effect-driven zone moves deliberately do NOT run it inline
// (zone_route.go), so a test that wants to see CR 704.5d has to reach
// the same boundary a real game reaches.
func settle(g *Game) {
	g.WithWriteLock(func() {
		g.runStateChecksLocked()
	})
}

// nowhere asserts the object is in no zone at all.
func nowhere(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if z := g.findCardZoneLocked(id); z != nil {
			t.Errorf("token should have ceased to exist, found it in %s", z.Kind)
		}
	})
}

// TestBouncedTokenCeasesToExist is the reported defect: Cyclonic Rift
// on a Bird token must not put a card in its controller's hand.
func TestBouncedTokenCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	handBefore := owner.Hand.Size()
	id := pushToken(g, owner, "Bird", "Token Creature — Bird")

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		// CR 111.7: it really does get to the hand first. Only the
		// state-based check takes it away.
		if !owner.Hand.Contains(id) {
			t.Fatalf("the bounce never reached the hand")
		}
	})

	settle(g)
	nowhere(t, g, id)
	if got := owner.Hand.Size(); got != handBefore {
		t.Errorf("hand size = %d, want %d — bouncing a token must not create card advantage", got, handBefore)
	}
}

// TestBouncedNontokenStaysInHand — the sweep is a token sweep. An
// ordinary Unsummon still returns a card.
func TestBouncedNontokenStaysInHand(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushIntrinsicPermanent(g, owner, "Bear", "Creature — Bear", nil, nil)

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	settle(g)

	if !owner.Hand.Contains(id) {
		t.Errorf("a bounced nontoken creature belongs in its owner's hand")
	}
}

// TestDestroyedTokenCeasesToExistAfterItDies — the graveyard leg. The
// token dies first (its LTB event goes out while it is in the
// graveyard, so "when this dies" payoffs see it), and only then does
// it stop existing.
func TestDestroyedTokenCeasesToExistAfterItDies(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushToken(g, owner, "Bird", "Token Creature — Bird")

	obs := &ltbZoneObserver{cardID: id}
	g.RegisterListener(obs)

	g.WithWriteLock(func() {
		if err := g.routeBattlefieldCardToOwnerGraveyardLocked(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !obs.saw {
		t.Fatalf("the dying token emitted no leaves-the-battlefield event")
	}
	if obs.zone != ZoneGraveyard {
		t.Errorf("at LTB time the token was in %q, want the graveyard — a dies trigger must see it there", obs.zone)
	}

	settle(g)
	nowhere(t, g, id)
	if owner.Graveyard.Contains(id) {
		t.Errorf("a dead token must not be left in the graveyard")
	}
}

// ltbZoneObserver records which zone the card was in at the moment
// its EventLTB fired — the window a dies trigger is harvested in.
type ltbZoneObserver struct {
	cardID uuid.UUID
	saw    bool
	zone   ZoneKind
}

func (o *ltbZoneObserver) OnEvent(g *Game, ev Event) {
	if ev.Kind != EventLTB || ev.CardID != o.cardID || o.saw {
		return
	}
	o.saw = true
	if z := g.findCardZoneLocked(o.cardID); z != nil {
		o.zone = z.Kind
	}
}

// TestSacrificedTokenCeasesToExist — the aristocrats path. The
// sacrifice announces itself (EventSacrifice) before the token goes.
func TestSacrificedTokenCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushToken(g, owner, "Servo", "Token Artifact Creature — Servo")

	if err := g.SacrificePermanent(owner.ID, id); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	if !hasEvent(g, EventSacrifice, id) {
		t.Errorf("no EventSacrifice — a token is sacrificed before it stops existing")
	}
	nowhere(t, g, id)
	if owner.Graveyard.Contains(id) {
		t.Errorf("a sacrificed token must not be left in the graveyard")
	}
}

// TestSacrificedTokenIsStillCountedByTheTurnTally — the payoff half
// of the sacrifice. "You sacrificed a Food this turn" is asked at the
// end step, long after CR 704.5d has taken the Food away, so the
// tally has to have recorded what it was at the sacrifice.
func TestSacrificedTokenIsStillCountedByTheTurnTally(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushToken(g, owner, "Food", "Token Artifact — Food")

	if err := g.SacrificePermanent(owner.ID, id); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	nowhere(t, g, id)

	g.WithWriteLock(func() {
		if n := g.SacrificedWithSubtypeThisTurn(owner.ID, "Food"); n != 1 {
			t.Errorf("sacrificed Food this turn = %d, want 1", n)
		}
		if n := g.SacrificedWithSubtypeThisTurn(owner.ID, "Clue"); n != 0 {
			t.Errorf("sacrificed Clue this turn = %d, want 0", n)
		}
		if n := g.SacrificedWithSubtypeThisTurn(g.Seats[1].ID, "Food"); n != 0 {
			t.Errorf("the tally is per player: opponent = %d, want 0", n)
		}
	})
}

// TestExiledTokenCeasesToExist — the exile leg (Swords to
// Plowshares, a blink that never blinks back).
func TestExiledTokenCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushToken(g, owner, "Soldier", "Token Creature — Soldier")

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		if !g.Exile.Contains(id) {
			t.Fatalf("the exile never landed")
		}
	})
	settle(g)

	nowhere(t, g, id)
	if g.Exile.Contains(id) {
		t.Errorf("an exiled token must not be left in exile")
	}
}

// TestTuckedTokenCeasesToExist — the library leg. Chaos Warp on a
// token shuffles nothing into the deck.
func TestTuckedTokenCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	libBefore := owner.Library.Size()
	id := pushToken(g, owner, "Treasure", "Token Artifact — Treasure")

	g.WithWriteLock(func() {
		if err := g.TuckToLibraryForEffect(id, false); err != nil {
			t.Fatalf("TuckToLibraryForEffect: %v", err)
		}
	})
	settle(g)

	nowhere(t, g, id)
	if got := owner.Library.Size(); got != libBefore {
		t.Errorf("library size = %d, want %d — a tucked token must not become a card in the deck", got, libBefore)
	}
}

// TestTokenInCommandZoneCeasesToExist — the last zone. Nothing routes
// a token here today (a token is never a commander), so the object is
// planted directly: the rule is "any zone other than the battlefield",
// and the sweep must not have a blind spot waiting for the first
// effect that finds one.
func TestTokenInCommandZoneCeasesToExist(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	tok := NewCard("Spirit", owner.ID)
	tok.TypeLine = "Token Creature — Spirit"
	owner.Command.PushTop(tok)

	settle(g)
	nowhere(t, g, tok.InstanceID)
}

// TestTokenSweepAnnouncesTheDeparture — the client's zone views are
// driven off the event stream, so the removal has to say so. Same
// shape spell copies use: a zone move to nowhere.
func TestTokenSweepAnnouncesTheDeparture(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushToken(g, owner, "Bird", "Token Creature — Bird")

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	settle(g)

	var announced bool
	for _, ev := range g.Events {
		if ev.Kind == EventZoneMove && ev.CardID == id && ev.OldZone == ZoneHand && ev.NewZone == "" {
			announced = true
		}
	}
	if !announced {
		t.Errorf("the token ceased to exist without an event saying so")
	}
}

// TestTokenSweepLeavesTheBattlefieldAlone — the obvious guard. A
// token on the battlefield is where a token belongs.
func TestTokenSweepLeavesTheBattlefieldAlone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := pushToken(g, owner, "Bird", "Token Creature — Bird")

	settle(g)

	if !g.Battlefield.Contains(id) {
		t.Errorf("CR 704.5d is about every zone EXCEPT the battlefield")
	}
}

// TestTokenSweepSpareNontokensInEveryZone — a card in a graveyard, an
// exile, a hand or a library is not a token and must survive the
// sweep. Cheap, and it is the assertion that would have caught a
// sweep written the other way round.
func TestTokenSweepSparesNontokensInEveryZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]

	gy := NewCard("Bear", owner.ID)
	gy.TypeLine = "Creature — Bear"
	owner.Graveyard.PushTop(gy)

	ex := NewCard("Bolt", owner.ID)
	ex.TypeLine = "Instant"
	g.Exile.PushTop(ex)

	settle(g)

	if !owner.Graveyard.Contains(gy.InstanceID) {
		t.Errorf("a creature card in the graveyard is not a token")
	}
	if !g.Exile.Contains(ex.InstanceID) {
		t.Errorf("an exiled instant is not a token")
	}
}
