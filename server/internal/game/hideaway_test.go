package game

import (
	"testing"

	"github.com/google/uuid"
)

// hideaway_test.go — CR 702.75, ADR 0091 (#1331), engine side: the
// FaceDownHidden exile, its viewer rule, the link, and playing the card
// from exile. The catalog's look-and-choose and the four proof cards
// are tested in cards/effects/hideaway_cards_test.go.

// hideawaySetup puts a hideaway permanent on seat 0's battlefield and
// a card on top of seat 0's library, and hides that card with the
// permanent. Returns the permanent and the hidden card.
func hideawaySetup(t *testing.T, g *Game, hidden Card) (source, cardID uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	source = pushTypedTestCard(g, Card{Name: "Test Heights", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	hidden.InstanceID = uuid.New()
	hidden.Owner = me.ID
	me.Library.PushTop(hidden)
	g.WithWriteLock(func() {
		ref := ObjectRefOf(*findBattlefieldCard(g, source))
		if paused, err := g.ExileHiddenForEffect(ref, me.ID, hidden.InstanceID); paused || err != nil {
			t.Fatalf("ExileHiddenForEffect = %v, %v", paused, err)
		}
	})
	return source, hidden.InstanceID
}

// CR 702.75a / CR 406.3: the card is exiled face down, and the one
// player who may look is the controller of the permanent that exiled
// it — not the table, and not by accident the owner.
func TestAHiddenCardIsFaceDownAndOnlyTheSourcesControllerMayLook(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source, id := hideawaySetup(t, g, Card{Name: "Hidden Bolt", TypeLine: "Instant", ManaCost: "{R}"})

	c, ok := hiddenZoneCard(g.Exile, id)
	if !ok {
		t.Fatal("the hidden card is not in exile")
	}
	if !c.FaceDown || c.FaceDownKind != FaceDownHidden {
		t.Errorf("face down = %v kind %q, want face down as %q", c.FaceDown, c.FaceDownKind, FaceDownHidden)
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Errorf("knowers: me=%v opp=%v, want only the source's controller", c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
	}
	var ref PermissionCardRef
	g.ReadSnapshot(func() { ref = ObjectRefOf(*findBattlefieldCard(g, source)) })
	if c.HiddenBy != ref {
		t.Errorf("HiddenBy = %+v, want the source object %+v", c.HiddenBy, ref)
	}
}

// CR 607.2a: "the exiled card" is what THAT object hid. A reference to
// the same card at another epoch — the land bounced and replayed — has
// hidden nothing.
func TestHiddenCardsAreLinkedToTheObjectThatHidThem(t *testing.T) {
	g := newActiveGame(t)
	source, id := hideawaySetup(t, g, Card{Name: "Hidden Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	g.ReadSnapshot(func() {
		ref := ObjectRefOf(*findBattlefieldCard(g, source))
		if got := g.HiddenCardsForEffect(ref); len(got) != 1 || got[0] != id {
			t.Errorf("HiddenCardsForEffect = %v, want [%s]", got, id)
		}
		later := ref
		later.Epoch++
		if got := g.HiddenCardsForEffect(later); len(got) != 0 {
			t.Errorf("a later incarnation of the source finds %v, want nothing", got)
		}
	})
}

// CR 702.75a names the CONTROLLER of the permanent, so the look follows
// the permanent when it changes hands; CR 406.3 keeps the player who
// has already looked.
func TestTheLookFollowsTheHideawayPermanentsController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source, id := hideawaySetup(t, g, Card{Name: "Hidden Bolt", TypeLine: "Instant", ManaCost: "{R}"})

	g.WithWriteLock(func() {
		findBattlefieldCard(g, source).Controller = opp.ID
		g.runStateChecksLocked()
	})
	c, _ := hiddenZoneCard(g.Exile, id)
	if !c.IsKnownTo(opp.ID) {
		t.Error("the new controller of the hideaway permanent may not look at the hidden card")
	}
	if !c.IsKnownTo(me.ID) {
		t.Error("CR 406.3: the player who already looked lost the look")
	}
}

// A hidden card that leaves exile carries no link: MoveCard drops it,
// so nothing can find it as "the exiled card" again.
func TestAHiddenCardLeavingExileDropsItsLink(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	_, id := hideawaySetup(t, g, Card{Name: "Hidden Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Exile, me.Hand, id); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	c, _ := hiddenZoneCard(me.Hand, id)
	if c.HiddenBy.ID != uuid.Nil || c.FaceDown {
		t.Errorf("the card in hand still carries HiddenBy %+v / face down %v", c.HiddenBy, c.FaceDown)
	}
}

// The link and the kind survive a snapshot.
func TestHiddenStateSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source, id := hideawaySetup(t, g, Card{Name: "Hidden Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	_, restored := roundTrip(t, g)
	c, ok := hiddenZoneCard(restored.Exile, id)
	if !ok || c.FaceDownKind != FaceDownHidden || c.HiddenBy.ID != source || !c.IsKnownTo(me.ID) {
		t.Fatalf("restored hidden card = %+v", c)
	}
}

// CR 406.3a: a card exiled face down that a player is allowed to play
// is turned face up as it is played. A free-play grant over a hidden
// spell casts it for {0}; over a hidden LAND it plays the land, which
// spends the land drop (CR 305.2a).
func TestAHiddenCardIsPlayedFromExileUnderAGrant(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	_, spell := hideawaySetup(t, g, Card{Name: "Hidden Sorcery", TypeLine: "Sorcery", ManaCost: "{5}{R}"})
	_, land := hideawaySetup(t, g, Card{Name: "Hidden Mountain", TypeLine: "Basic Land — Mountain"})

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{spell, land} {
			if !g.GrantCastPermissionOverCardForEffect(id, CastPermission{Player: me.ID, Zone: ZoneExile, Cost: "{0}", Timing: TimingFlash}) {
				t.Fatalf("grant over %s refused", id)
			}
		}
	})
	if err := g.CastSpell(me.ID, spell, CastSpellParams{FromZone: string(ZoneExile), Strict: true}); err != nil {
		t.Fatalf("cast the hidden sorcery for {0}: %v", err)
	}
	onStack, ok := hiddenZoneCard(g.Stack, spell)
	if !ok || onStack.FaceDown || onStack.HiddenBy.ID != uuid.Nil {
		t.Errorf("on the stack: present=%v face down=%v link=%+v — want a face-up spell with no link", ok, onStack.FaceDown, onStack.HiddenBy)
	}
	for i := 0; i < 8 && g.Stack.Contains(spell); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	lands := g.LandsPlayedThisTurnFor(me.ID)
	if err := g.CastSpell(me.ID, land, CastSpellParams{FromZone: string(ZoneExile)}); err != nil {
		t.Fatalf("play the hidden land: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the hidden land is not on the battlefield")
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != lands+1 {
		t.Errorf("lands played = %d, want %d — playing a land from exile is a land play", got, lands+1)
	}
}

// hiddenZoneCard reads one card out of a zone by instance ID.
func hiddenZoneCard(z *Zone, id uuid.UUID) (Card, bool) {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return Card{}, false
}
