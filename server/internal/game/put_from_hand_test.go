package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// put_from_hand_test.go — PutFromHandOntoBattlefieldForEffect (#654):
// the hand-to-battlefield move behind "you may put a land card from
// your hand onto the battlefield".
//
// What is being pinned is that it is an ORDINARY battlefield entry —
// the CR 614 pipeline, the ETB event and the catalog hook, the same
// four steps the search, reanimation and exile-return paths take —
// and that it is NOT a land play.

// handCardFor seeds one card into a player's hand and returns its ID.
func handCardFor(p *Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// battlefieldCardForTest returns a copy of a battlefield card by ID.
func battlefieldCardForTest(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return Card{}
}

// The base case: the land leaves hand for the battlefield under its
// owner's control, untapped, known to the whole table — and the turn's
// land drop is still there, because a land an effect PUTS onto the
// battlefield was never PLAYED (CR 305.4).
func TestPutFromHandOntoBattlefieldEntersAndIsNotALandDrop(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := handCardFor(me, "Forest", "Basic Land — Forest")

	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutFromHandOntoBattlefieldForEffect(land, HandEntryOptions{})
	})
	if err != nil {
		t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
	}
	if entered != land {
		t.Errorf("entered as %s, want the hand instance %s — only an exile return mints a new ID", entered, land)
	}
	if me.Hand.Contains(land) {
		t.Error("the land is still in hand")
	}
	c := battlefieldCardForTest(t, g, land)
	if c.Controller != me.ID {
		t.Errorf("controller %s, want the owner %s", c.Controller, me.ID)
	}
	if c.Tapped {
		t.Error("the land entered tapped with nothing saying so")
	}
	for _, seat := range g.Seats {
		if !c.IsKnownTo(seat.ID) {
			t.Errorf("seat %s does not know the land: it left a hidden zone for a public one", seat.Name)
		}
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d, want 0 — putting a land onto the battlefield is not playing one (CR 305.4)", n)
	}
}

// The move is an entry, so it emits the zone move and EventETB the
// rest of the engine watches — a landfall trigger and an ETB trigger
// both hang off these.
func TestPutFromHandOntoBattlefieldEmitsTheEntryEvents(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	rock := handCardFor(me, "Sol Ring", "Artifact")
	before := len(g.Events)

	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(rock, HandEntryOptions{}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})

	var sawMove, sawETB bool
	for _, ev := range g.Events[before:] {
		switch ev.Kind {
		case EventZoneMove:
			if ev.CardID == rock && ev.OldZone == ZoneHand && ev.NewZone == ZoneBattlefield && ev.Actor == me.ID {
				sawMove = true
			}
		case EventETB:
			if ev.CardID == rock && ev.Actor == me.ID {
				sawETB = true
			}
		case EventTapCard:
			t.Error("nothing was tapped, so no tap event may be emitted")
		}
	}
	if !sawMove {
		t.Error("no hand → battlefield EventZoneMove")
	}
	if !sawETB {
		t.Error("no EventETB: an ETB trigger would never fire")
	}
}

// A card's own "enters tapped" replacement is a CR 614 replacement on
// this move like any other, and the permanent ARRIVES tapped — it is
// never tapped after the fact, so no tap event fires. This is the
// engine-side twin of a land declaring SelfEntersTapped() in the
// catalog.
func TestPutFromHandOntoBattlefieldRunsTheEntryPipeline(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := handCardFor(me, "Jungle Hollow", "Land")
	before := len(g.Events)

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield && ev.CardID == land
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.EntersTapped = true
				ev.EntersWithCounters = map[string]int{"+1/+1": 1}
				return nil
			},
			Label: "Test: this land enters tapped with a counter",
		})
		if _, err := g.PutFromHandOntoBattlefieldForEffect(land, HandEntryOptions{}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})

	c := battlefieldCardForTest(t, g, land)
	if !c.Tapped {
		t.Error("the entry replacement said tapped and the land arrived untapped")
	}
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("entered with %d +1/+1 counters, want 1", c.Counters["+1/+1"])
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventTapCard && ev.CardID == land {
			t.Error("the land was tapped after entering; it must ARRIVE tapped")
		}
	}
}

// The effect's own "onto the battlefield tapped" clause is seeded
// onto the same event, so it settles in one place with whatever CR
// 614 adds — and again without a tap event.
func TestPutFromHandOntoBattlefieldHonoursTheTappedOption(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := handCardFor(me, "Forest", "Basic Land — Forest")
	before := len(g.Events)

	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(land, HandEntryOptions{Tapped: true}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})

	if !battlefieldCardForTest(t, g, land).Tapped {
		t.Error("the effect said tapped and the land arrived untapped")
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventTapCard && ev.CardID == land {
			t.Error("tapped on entry must not emit a tap event")
		}
	}
}

// "Put it onto the battlefield UNDER YOUR CONTROL" when the card is
// somebody else's: the controller is stamped BEFORE the pipeline
// runs, which is what an "is this permanent mine?" replacement reads.
func TestPutFromHandOntoBattlefieldHonoursAnExplicitController(t *testing.T) {
	g := newActiveGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	card := handCardFor(owner, "Bear", "Creature — Bear")

	var sawController uuid.UUID
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, gg *Game, _ *Card) bool {
				if ev.Kind != RepEventMove || ev.CardID != card {
					return false
				}
				if c, ok := gg.LookupCardForEffect(card); ok {
					sawController = c.Controller
				}
				return false
			},
			Label: "Test: observe the entering permanent's controller",
		})
		if _, err := g.PutFromHandOntoBattlefieldForEffect(card, HandEntryOptions{Controller: thief.ID}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})

	if sawController != thief.ID {
		t.Errorf("the pipeline saw controller %s, want %s — it is stamped before the replacements run", sawController, thief.ID)
	}
	if got := battlefieldCardForTest(t, g, card).Controller; got != thief.ID {
		t.Errorf("controller %s, want %s", got, thief.ID)
	}
	if got := battlefieldCardForTest(t, g, card).Owner; got != owner.ID {
		t.Errorf("owner %s, want %s — control changes, ownership does not", got, owner.ID)
	}
}

// A replacement that cancels the move leaves the card in hand, and
// says so with uuid.Nil rather than an error: nothing went wrong, the
// permanent just did not enter.
func TestPutFromHandOntoBattlefieldCanceledLeavesTheCardInHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	land := handCardFor(me, "Forest", "Basic Land — Forest")

	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == land
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Canceled = true
				return nil
			},
			Label: "Test: this entry is replaced with nothing",
		})
		entered, err = g.PutFromHandOntoBattlefieldForEffect(land, HandEntryOptions{})
	})

	if err != nil {
		t.Fatalf("a canceled entry is not an error: %v", err)
	}
	if entered != uuid.Nil {
		t.Errorf("returned %s, want uuid.Nil", entered)
	}
	if !me.Hand.Contains(land) {
		t.Error("the canceled land left the hand anyway")
	}
}

// It is a HAND move. A card anywhere else is refused rather than
// silently ripped out of the zone it is in.
func TestPutFromHandOntoBattlefieldRefusesACardNotInHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	onBattlefield := permanentFor(g, me, "Bear", "Creature — Bear", "{1}{G}")
	inGraveyard := uuid.New()
	me.Graveyard.PushTop(Card{InstanceID: inGraveyard, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID})

	for _, tc := range []struct {
		name string
		id   uuid.UUID
	}{
		{"on the battlefield", onBattlefield},
		{"in a graveyard", inGraveyard},
		{"nowhere at all", uuid.New()},
	} {
		var err error
		g.WithWriteLock(func() {
			_, err = g.PutFromHandOntoBattlefieldForEffect(tc.id, HandEntryOptions{})
		})
		if !errors.Is(err, ErrCardNotFound) {
			t.Errorf("%s: err = %v, want ErrCardNotFound", tc.name, err)
		}
	}
	if !g.Battlefield.Contains(onBattlefield) || !me.Graveyard.Contains(inGraveyard) {
		t.Error("a refused put moved something anyway")
	}
}

// Only a permanent card can become a permanent (CR 110.4). An instant
// stays in hand.
func TestPutFromHandOntoBattlefieldRefusesANonpermanentCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bolt := handCardFor(me, "Lightning Bolt", "Instant")

	var err error
	g.WithWriteLock(func() {
		_, err = g.PutFromHandOntoBattlefieldForEffect(bolt, HandEntryOptions{})
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if !me.Hand.Contains(bolt) {
		t.Error("the instant left the hand")
	}
}
