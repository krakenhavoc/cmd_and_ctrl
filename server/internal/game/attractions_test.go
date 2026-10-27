package game

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// attractions_test.go — ADR 0207: the Attraction deck, opening an
// Attraction (CR 701.51), rolling to visit (CR 701.52, 717.4) and the
// junkyard (CR 717.6).

// attrCard is an Attraction card with the given lights.
func attrCard(name string, lights ...int) Card {
	c := NewCard(name, uuid.Nil)
	c.TypeLine = "Artifact — Attraction"
	c.AstrotoriumBack = true
	c.AttractionLights = lights
	return c
}

// attrDeck is n Attraction cards named "Attraction 1".."Attraction n".
func attrDeck(n int) []Card {
	out := make([]Card, n)
	for i := range out {
		out[i] = attrCard(fmt.Sprintf("Attraction %d", i+1), i%6+1)
	}
	return out
}

// newAttractionGame starts a two-seat game in which seat 0 plays with
// `attractions` as its Attraction deck, with the mulligans kept.
func newAttractionGame(t *testing.T, seed uint64, attractions ...Card) *Game {
	t.Helper()
	g := NewGame()
	for i := range 2 {
		deck := buildTestDeck(fmt.Sprintf("Commander %d", i+1))
		if i == 0 {
			deck = append(deck, attractions...)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// putAttraction puts an Attraction onto the battlefield under `owner`.
func putAttraction(g *Game, owner uuid.UUID, name string, lights ...int) uuid.UUID {
	c := attrCard(name, lights...)
	c.Owner, c.Controller = owner, owner
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func attrEvents(g *Game, from int, kind EventKind) []Event {
	var out []Event
	for _, ev := range g.Events[from:] {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

func attrNames(cards []Card) []string {
	out := make([]string, len(cards))
	for i, c := range cards {
		out[i] = c.Name
	}
	return out
}

// CR 717.2: Attraction cards begin the game in a face-down Attraction
// deck in the command zone, not in the library, not among the
// commanders, and out of reach of every lookup by ID.
func TestAttractionDeckSeatsFaceDownInTheCommandZone(t *testing.T) {
	g := newAttractionGame(t, 1, attrDeck(10)...)
	me, opp := g.Seats[0], g.Seats[1]
	deck := me.AttractionDeckCards()
	if len(deck) != 10 {
		t.Fatalf("Attraction deck holds %d, want 10", len(deck))
	}
	if me.AttractionDeck.Kind != ZoneCommand || me.AttractionDeck.Owner != me.ID {
		t.Errorf("container is %s owned by %v", me.AttractionDeck.Kind, me.AttractionDeck.Owner)
	}
	if n := me.Library.Size() + me.Hand.Size() + me.Command.Size(); n != 100 {
		t.Errorf("library + hand + command = %d, want the deck's 100", n)
	}
	for _, c := range deck {
		if c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
			t.Errorf("%s is known to a player in a face-down deck", c.Name)
		}
		if _, ok := g.LookupCardForEffect(c.InstanceID); ok {
			t.Errorf("%s is reachable by LookupCardForEffect", c.Name)
		}
	}
	if opp.AttractionDeckCards() != nil {
		t.Error("a seat with no Attractions has an Attraction deck")
	}
}

// CR 103.3a: the Attraction deck is shuffled before the game begins, on
// its own stream, so the library shuffles exactly as it would without it.
func TestAttractionDeckIsShuffledApartFromTheLibrary(t *testing.T) {
	seated := attrNames(attrDeck(10))
	with := newAttractionGame(t, 7, attrDeck(10)...)
	without := newAttractionGame(t, 7)
	if got := attrNames(with.Seats[0].AttractionDeckCards()); slices.Equal(got, seated) {
		t.Errorf("the Attraction deck is in seating order %v: it was not shuffled", got)
	}
	if !slices.Equal(attrNames(with.Seats[0].Library.Cards), attrNames(without.Seats[0].Library.Cards)) {
		t.Error("seating an Attraction deck changed the library's shuffle")
	}
}

// CR 701.51b: opening puts the TOP card onto the battlefield under the
// player's control, face up; CR 701.51c announces it once per open.
func TestOpenAnAttractionPutsTheTopCardOntoTheBattlefield(t *testing.T) {
	g := newAttractionGame(t, 1, attrDeck(10)...)
	me, opp := g.Seats[0], g.Seats[1]
	top := me.AttractionDeck.Cards[len(me.AttractionDeck.Cards)-1].InstanceID
	second := me.AttractionDeck.Cards[len(me.AttractionDeck.Cards)-2].InstanceID
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsForEffect(me.ID, uuid.Nil, 2); err != nil {
			t.Fatal(err)
		}
	})
	var opened []uuid.UUID
	for _, ev := range attrEvents(g, before, EventAttractionOpened) {
		opened = append(opened, ev.CardID)
	}
	if !slices.Equal(opened, []uuid.UUID{top, second}) {
		t.Fatalf("opened %v, want %v then %v", opened, top, second)
	}
	for _, id := range opened {
		c := findBattlefieldCard(g, id)
		if c == nil || c.Controller != me.ID || c.Owner != me.ID {
			t.Fatalf("opened Attraction %v is not on the battlefield under its owner: %+v", id, c)
		}
		if !c.IsKnownTo(opp.ID) {
			t.Error("an opened Attraction is not face up to the table")
		}
	}
	evs := attrEvents(g, before, EventAttractionOpened)
	if len(evs) != 2 || evs[0].CardID != top || evs[1].CardID != second || evs[0].Actor != me.ID {
		t.Errorf("open events %+v", evs)
	}
	if len(me.AttractionDeckCards()) != 8 {
		t.Errorf("Attraction deck holds %d, want 8", len(me.AttractionDeckCards()))
	}
	if g.attractionEntering != uuid.Nil {
		t.Error("the open's index window was left open")
	}
}

// CR 701.51a: a player with no Attraction deck, or an empty one, opens
// nothing, and that is not an error.
func TestOpenAnAttractionWithNoDeckOpensNothing(t *testing.T) {
	g := newAttractionGame(t, 1)
	opp := g.Seats[1]
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsForEffect(opp.ID, uuid.Nil, 1); err != nil {
			t.Fatal(err)
		}
	})
	if evs := attrEvents(g, before, EventAttractionOpened); len(evs) != 0 {
		t.Errorf("opened %d with no deck", len(evs))
	}
}

// CR 701.52a: a d6 (an ordinary die roll), then each Attraction the
// player controls with that number lit up is visited, and nothing else.
func TestRollToVisitVisitsTheAttractionsLitForTheResult(t *testing.T) {
	g := newAttractionGame(t, 1)
	me, opp := g.Seats[0], g.Seats[1]
	low := putAttraction(g, me.ID, "Low", 1, 2, 3)
	high := putAttraction(g, me.ID, "High", 4, 5, 6)
	theirs := putAttraction(g, opp.ID, "Theirs", 1, 2, 3, 4, 5, 6)
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.RollToVisitForEffect(me.ID, uuid.Nil); err != nil {
			t.Fatal(err)
		}
	})
	rolls := attrEvents(g, before, EventRollDie)
	result := 0
	if len(rolls) == 1 {
		result = rolls[0].Amount
	}
	if len(rolls) != 1 || rolls[0].Sides != 6 || rolls[0].Amount != result || rolls[0].Actor != me.ID {
		t.Fatalf("die rolls %+v for result %d", rolls, result)
	}
	visits := attrEvents(g, before, EventRollToVisit)
	if len(visits) != 1 || visits[0].Amount != result {
		t.Fatalf("roll-to-visit events %+v", visits)
	}
	want := low
	if result >= 4 {
		want = high
	}
	visited := attrEvents(g, before, EventVisitAttraction)
	if len(visited) != 1 || visited[0].CardID != want || visited[0].Amount != result {
		t.Fatalf("visited %+v on a %d, want only %v", visited, result, want)
	}
	if visited[0].CardID == theirs {
		t.Error("an opponent's Attraction was visited")
	}
}

// CR 717.4 / 703.4g: as the precombat main phase begins, the active
// player rolls to visit if, and only if, they control an Attraction.
func TestRollToVisitIsATurnBasedActionOfThePrecombatMainPhase(t *testing.T) {
	g := newAttractionGame(t, 1)
	before := len(g.Events)
	advanceTo(t, g, StepPrecombatMain)
	if evs := attrEvents(g, before, EventRollToVisit); len(evs) != 0 {
		t.Fatalf("rolled with no Attractions: %+v", evs)
	}
	g2 := newAttractionGame(t, 1)
	active := g2.Seats[g2.Turn.ActiveSeat]
	putAttraction(g2, active.ID, "Booth", 1)
	before = len(g2.Events)
	advanceTo(t, g2, StepPrecombatMain)
	evs := attrEvents(g2, before, EventRollToVisit)
	if len(evs) != 1 || evs[0].Actor != active.ID {
		t.Fatalf("roll-to-visit events %+v, want one by the active player", evs)
	}
	began := attrEvents(g2, before, EventBeginPrecombatMain)
	if len(began) == 0 || began[0].Seq < evs[0].Seq {
		t.Error("the roll did not come before the precombat main phase's beginning triggers")
	}
}

// CR 717.6: an Astrotorium-backed card bound for a graveyard, a hand or
// a library goes to its owner's junkyard instead, face up; exile is left
// alone; a token is not a card and has no back.
func TestTheJunkyardReplacement(t *testing.T) {
	type mover func(g *Game, id uuid.UUID, owner uuid.UUID) error
	destroy := func(g *Game, id, _ uuid.UUID) error {
		var err error
		g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
		return err
	}
	bounce := func(g *Game, id, _ uuid.UUID) error {
		var err error
		g.WithWriteLock(func() { err = g.BounceToHandForEffect(id) })
		return err
	}
	tuck := func(g *Game, id, owner uuid.UUID) error {
		return g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneLibrary, Owner: owner}, id)
	}
	for _, tc := range []struct {
		name string
		move mover
	}{{"destroyed", destroy}, {"bounced", bounce}, {"put into a library", tuck}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newAttractionGame(t, 1)
			me, opp := g.Seats[0], g.Seats[1]
			id := putAttraction(g, me.ID, "Booth", 2, 6)
			// Under an opponent's control: the junkyard is the OWNER's.
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Controller = opp.ID
				}
			}
			if err := tc.move(g, id, me.ID); err != nil {
				t.Fatalf("move: %v", err)
			}
			if !slices.ContainsFunc(me.JunkyardCards(), func(c Card) bool { return c.InstanceID == id }) {
				t.Fatalf("not in its owner's junkyard; graveyard %v hand %v library %v",
					me.Graveyard.Contains(id), me.Hand.Contains(id), me.Library.Contains(id))
			}
			if me.Command.Contains(id) || opp.Junkyard != nil {
				t.Error("the card landed among the commanders or in the controller's junkyard")
			}
			if me.Junkyard.Kind != ZoneCommand {
				t.Errorf("junkyard kind %s", me.Junkyard.Kind)
			}
			c := me.JunkyardCards()[0]
			if !c.IsKnownTo(opp.ID) || !c.IsKnownTo(me.ID) {
				t.Error("a junkyard card is not face up to the table")
			}
		})
	}
	t.Run("exiled", func(t *testing.T) {
		g := newAttractionGame(t, 1)
		me := g.Seats[0]
		id := putAttraction(g, me.ID, "Booth", 2, 6)
		if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneExile}, id); err != nil {
			t.Fatal(err)
		}
		if !g.Exile.Contains(id) || me.Junkyard != nil {
			t.Error("an exiled Attraction was not left in exile (CR 717.6 spares exile)")
		}
	})
	t.Run("a token", func(t *testing.T) {
		g := newAttractionGame(t, 1)
		me := g.Seats[0]
		c := attrCard("Token Booth", 2, 6)
		c.TypeLine = "Token Artifact — Attraction"
		c.Owner, c.Controller = me.ID, me.ID
		g.Battlefield.PushTop(c)
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(c.InstanceID); err != nil {
				t.Fatal(err)
			}
		})
		if me.Junkyard != nil {
			t.Error("a token went to the junkyard")
		}
	})
}

// A restore point while the Attraction deck, an open Attraction and the
// junkyard are all live: the deck's order, the lights and the junkyard
// survive capture → JSON → restore, and the restored table opens the
// same card next.
func TestAttractionsSurviveARestore(t *testing.T) {
	g := newAttractionGame(t, 3, attrDeck(10)...)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsForEffect(me.ID, uuid.Nil, 2); err != nil {
			t.Fatal(err)
		}
	})
	opened := attrEvents(g, 0, EventAttractionOpened)
	if len(opened) != 2 {
		t.Fatalf("opened %d", len(opened))
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(opened[0].CardID); err != nil {
			t.Fatal(err)
		}
	})
	_, restored := roundTrip(t, g)
	rme := restored.Seats[0]
	if !slices.Equal(attrNames(rme.AttractionDeckCards()), attrNames(me.AttractionDeckCards())) {
		t.Fatalf("deck order %v, want %v", attrNames(rme.AttractionDeckCards()), attrNames(me.AttractionDeckCards()))
	}
	if rme.AttractionDeck.Kind != ZoneCommand || len(rme.JunkyardCards()) != 1 || rme.Junkyard.Kind != ZoneCommand {
		t.Fatalf("piles after restore: deck %v junkyard %v", rme.AttractionDeck, rme.Junkyard)
	}
	live := findBattlefieldCard(restored, opened[1].CardID)
	if live == nil || !live.AstrotoriumBack || len(live.AttractionLights) == 0 {
		t.Fatalf("the open Attraction lost its back or its lights: %+v", live)
	}
	top := rme.AttractionDeck.Cards[len(rme.AttractionDeck.Cards)-1].InstanceID
	restored.WithWriteLock(func() {
		if err := restored.OpenAttractionsForEffect(rme.ID, uuid.Nil, 1); err != nil {
			t.Fatal(err)
		}
	})
	if findBattlefieldCard(restored, top) == nil {
		t.Error("the restored table did not open its top card")
	}
	// And an undo across an open puts the card back on top.
	snap := g.Clone()
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsForEffect(me.ID, uuid.Nil, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RestoreFrom(snap)
	if len(g.Seats[0].AttractionDeckCards()) != 8 {
		t.Errorf("undo left %d cards in the Attraction deck, want 8", len(g.Seats[0].AttractionDeckCards()))
	}
}

// CR 727.2: a restart involves nontraditional cards too, and CR 717.2
// starts each one in its owner's Attraction deck: the opened one on the
// battlefield, the one in the junkyard and the ones still in the deck all
// begin the new game there, shuffled, as new objects, and the library
// holds none of them.
func TestARestartRebuildsTheAttractionDeck(t *testing.T) {
	g := newAttractionGame(t, 5, attrDeck(10)...)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsForEffect(me.ID, uuid.Nil, 2); err != nil {
			t.Fatal(err)
		}
	})
	opened := attrEvents(g, 0, EventAttractionOpened)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(opened[0].CardID); err != nil {
			t.Fatal(err)
		}
	})
	if len(me.AttractionDeckCards()) != 8 || len(me.JunkyardCards()) != 1 {
		t.Fatalf("before the restart: deck %d junkyard %d", len(me.AttractionDeckCards()), len(me.JunkyardCards()))
	}
	old := map[uuid.UUID]bool{}
	for _, c := range me.AttractionDeckCards() {
		old[c.InstanceID] = true
	}
	g.WithWriteLock(func() { g.RestartTheGameForEffect(me.ID, uuid.Nil, nil) })
	g.SettleResolution()
	if g.RestartPending != nil || g.State != StateActive {
		t.Fatalf("the restart did not run: pending %v state %s", g.RestartPending, g.State)
	}
	nme := g.Seats[0]
	deck := nme.AttractionDeckCards()
	if len(deck) != 10 {
		t.Fatalf("the new game's Attraction deck holds %d, want all 10", len(deck))
	}
	if nme.JunkyardCards() != nil {
		t.Errorf("the new game starts with a junkyard of %d", len(nme.JunkyardCards()))
	}
	for _, c := range deck {
		if old[c.InstanceID] {
			t.Errorf("%s kept its old object ID across the restart", c.Name)
		}
		if !c.AstrotoriumBack || len(c.AttractionLights) == 0 {
			t.Errorf("%s lost its back or its lights", c.Name)
		}
	}
	for _, z := range []*Zone{nme.Library, nme.Hand, nme.Command} {
		for _, c := range z.Cards {
			if c.AstrotoriumBack {
				t.Errorf("%s began the new game in the %s", c.Name, z.Kind)
			}
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.AstrotoriumBack {
			t.Errorf("%s is still on the battlefield", c.Name)
		}
	}
}

// CR 800.4a: a player who leaves takes everything they own with them,
// the Attraction deck and the junkyard included.
func TestALeavingPlayerTakesTheirAttractionPiles(t *testing.T) {
	g := NewGame()
	for i := range 3 {
		deck := buildTestDeck(fmt.Sprintf("Commander %d", i+1))
		if i == 1 {
			deck = append(deck, attrDeck(10)...)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(9, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	leaver := g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.OpenAttractionsForEffect(leaver.ID, uuid.Nil, 2); err != nil {
			t.Fatal(err)
		}
	})
	opened := attrEvents(g, 0, EventAttractionOpened)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(opened[0].CardID); err != nil {
			t.Fatal(err)
		}
	})
	if len(leaver.AttractionDeckCards()) != 8 || len(leaver.JunkyardCards()) != 1 {
		t.Fatalf("before leaving: deck %d junkyard %d", len(leaver.AttractionDeckCards()), len(leaver.JunkyardCards()))
	}
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := len(leaver.AttractionDeckCards()) + len(leaver.JunkyardCards()); n != 0 {
		t.Errorf("the leaver's Attraction piles still hold %d cards", n)
	}
	if findBattlefieldCard(g, opened[1].CardID) != nil {
		t.Error("the leaver's opened Attraction is still on the battlefield")
	}
}
