package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// entry_gate_test.go — #1985: "can't enter the battlefield" (CR 101.2)
// asked at the top of every pass of the CR 614 window, so every entry
// path stops: a land play, a library search, a reanimation, an exile
// return, a hand put, a created token and a resolving permanent spell.
// The cards that declare it are pinned in cards/effects; the enumerator
// half is in internal/legal.

const testEntryBanOracle = "test-entry-ban"

// testLandsCantEnter is Worms of the Earth's second line, spelled out
// so the game package can test it without importing effects.
var testLandsCantEnter = EntryRestriction{
	Label:   "Lands can't enter the battlefield.",
	Forbids: func(q EntryQuery) bool { return q.WouldBe().IsLand() },
}

// stubEntryRestrictions wires CatalogEntryRestrictions for one oracle ID
// and restores the previous hook at the end of the test.
func stubEntryRestrictions(t *testing.T, oracleID string, rules []EntryRestriction) {
	t.Helper()
	prev := CatalogEntryRestrictions
	CatalogEntryRestrictions = func(id string) []EntryRestriction {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogEntryRestrictions = prev })
}

// seedEntryBan puts "Test Worms" on the battlefield with `rule`.
func seedEntryBan(t *testing.T, g *Game, controller uuid.UUID, rule EntryRestriction) uuid.UUID {
	t.Helper()
	stubEntryRestrictions(t, testEntryBanOracle, []EntryRestriction{rule})
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID: id,
			Name:       "Test Worms",
			TypeLine:   "Enchantment",
			OracleID:   testEntryBanOracle,
			Owner:      controller,
			Controller: controller,
		})
	})
	return id
}

// refusedEntry reports whether an EventEntryRefused names `card`.
func refusedEntry(g *Game, card uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventEntryRefused && ev.CardID == card {
				found = true
			}
		}
	})
	return found
}

func enteredBattlefield(g *Game, id uuid.UUID) bool {
	in := false
	g.ReadSnapshot(func() { in = g.Battlefield.Contains(id) })
	return in
}

// A land PLAY is refused before it is made, as a "can't play" carrying
// the entry clause, and spends no land drop.
func TestLandsCantEnterRefusesTheLandPlay(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	seedEntryBan(t, g, active.ID, testLandsCantEnter)
	handBefore := len(active.Hand.Cards)
	err := dropLandFromHand(t, g, active)
	assertCantPlayLand(t, err, "Lands can't enter the battlefield. — Test Worms")
	if len(active.Hand.Cards) != handBefore {
		t.Errorf("a refused land play moved the card: hand %d, want %d", len(active.Hand.Cards), handBefore)
	}
	if g.LandsPlayedThisTurnFor(active.ID) != 0 {
		t.Error("a refused land play spent a land drop")
	}
	// The gate the view and the enumerator ask under a read lock.
	if err := g.LandPlayGateFor(active.ID, Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, ZoneHand); !errors.Is(err, ErrCantPlayLand) {
		t.Errorf("LandPlayGateFor = %v, want ErrCantPlayLand", err)
	}
}

// A fetched land stays in the library; the search still completes, and a
// creature the same search finds still enters.
func TestLandsCantEnterKeepsAFetchedLandInTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedEntryBan(t, g, me.ID, testLandsCantEnter)
	forest := topOfLibraryFor(me, "Forest", "Basic Land — Forest")
	bear := topOfLibraryFor(me, "Bear", "Creature — Bear")
	g.mu.Lock()
	err := g.SearchLibraryThenForEffect(SearchLibrarySpec{
		Player: me.ID,
		Pred:   func(c Card) bool { return c.InstanceID == forest || c.InstanceID == bear },
		Dest:   ZoneBattlefield,
		Limit:  2,
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	takeSearchPrompts(t, g)
	if enteredBattlefield(g, forest) || !me.Library.Contains(forest) {
		t.Error("the fetched land left the library")
	}
	if !refusedEntry(g, forest) {
		t.Error("no EventEntryRefused for the fetched land")
	}
	if !enteredBattlefield(g, bear) {
		t.Error("the creature the same search found did not enter")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the search left %d prompts open", len(g.PendingChoices))
	}
}

// A reanimated land stays in the graveyard, a flickered land stays in
// exile, and a land put from hand stays in hand.
func TestLandsCantEnterKeepsTheLandWhereItWas(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedEntryBan(t, g, me.ID, testLandsCantEnter)

	grave := NewCard("Forest", me.ID)
	grave.TypeLine = "Basic Land — Forest"
	me.Graveyard.PushTop(grave)
	exiled := NewCard("Island", me.ID)
	exiled.TypeLine = "Basic Land — Island"
	g.Exile.PushTop(exiled)
	hand := handCardFor(me, "Swamp", "Basic Land — Swamp")

	var entered uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(grave.InstanceID, ZoneBattlefield); err != nil {
			t.Errorf("reanimate: %v", err)
		}
		id, err := g.ReturnFromExileToBattlefieldForEffect(exiled.InstanceID, me.ID, false)
		if err != nil {
			t.Errorf("exile return: %v", err)
		}
		entered = id
		if _, err := g.PutFromHandOntoBattlefieldForEffect(hand, HandEntryOptions{}); err != nil {
			t.Errorf("hand put: %v", err)
		}
	})
	if !me.Graveyard.Contains(grave.InstanceID) {
		t.Error("the reanimated land left the graveyard")
	}
	if entered != uuid.Nil || !g.Exile.Contains(exiled.InstanceID) {
		t.Errorf("the exiled land returned (entered %v)", entered)
	}
	if !me.Hand.Contains(hand) {
		t.Error("the land put from hand left the hand")
	}
	for _, id := range []uuid.UUID{grave.InstanceID, exiled.InstanceID, hand} {
		if !refusedEntry(g, id) {
			t.Errorf("no EventEntryRefused for %v", id)
		}
	}
}

// A land token is never created (CR 111.1); a creature token from the
// same kind of instruction is.
func TestLandsCantEnterCreatesNoLandToken(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedEntryBan(t, g, me.ID, testLandsCantEnter)
	before := len(g.Battlefield.Cards)
	var lands, bears []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		lands, err = g.CreateTokensForEffect(me.ID, Card{Name: "Land Token", TypeLine: "Token Land — Forest"}, 2, TokenEntryOptions{})
		if err != nil {
			t.Errorf("land tokens: %v", err)
		}
		bears, err = g.CreateTokensForEffect(me.ID, Card{Name: "Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2}, 1, TokenEntryOptions{})
		if err != nil {
			t.Errorf("creature token: %v", err)
		}
	})
	if len(lands) != 0 {
		t.Errorf("%d land tokens were created, want none", len(lands))
	}
	if len(bears) != 1 {
		t.Fatalf("%d creature tokens were created, want 1", len(bears))
	}
	if got := len(g.Battlefield.Cards); got != before+1 {
		t.Errorf("battlefield %d → %d, want +1 (the creature token)", before, got)
	}
	if len(g.enteringTokens) != 0 {
		t.Errorf("%d staged tokens left behind", len(g.enteringTokens))
	}
}

// The entering object is judged as the permanent it would be (CR
// 614.12's look-ahead): a land manifested face down is a typeless 2/2
// creature (CR 708.2) and enters, and a creature told to enter as a copy
// of a land is a land and does not.
func TestLandsCantEnterJudgesThePermanentItWouldBe(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedEntryBan(t, g, me.ID, testLandsCantEnter)
	land := topOfLibraryFor(me, "Forest", "Basic Land — Forest")
	var manifested uuid.UUID
	g.WithWriteLock(func() {
		id, err := g.ManifestForEffect(me.ID)
		if err != nil {
			t.Errorf("manifest: %v", err)
		}
		manifested = id
	})
	if manifested == uuid.Nil || !enteredBattlefield(g, manifested) {
		t.Errorf("a land manifested face down did not enter (got %v, top was %v)", manifested, land)
	}

	bear := handCardFor(me, "Bear", "Creature — Bear")
	forest := Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
	copyOf := CopiableValuesOf(forest)
	var out *ReplacementEvent
	ev := &ReplacementEvent{
		Kind: RepEventMove, CardID: bear, OldZone: ZoneHand, NewZone: ZoneBattlefield,
		Actor: me.ID, EntersAsCopyOf: &copyOf,
	}
	g.WithWriteLock(func() {
		out, _ = g.applyReplacementsLocked(ev)
		g.clearReplacementEventLocked(ev.ID)
	})
	if out != nil || !ev.Canceled || !ev.entryRefused {
		t.Errorf("a creature entering as a copy of a land was let in (out %v, canceled %v)", out, ev.Canceled)
	}
	// The same creature entering as itself is not a land.
	plain := &ReplacementEvent{Kind: RepEventMove, CardID: bear, OldZone: ZoneHand, NewZone: ZoneBattlefield, Actor: me.ID}
	g.WithWriteLock(func() {
		out, _ = g.applyReplacementsLocked(plain)
		g.clearReplacementEventLocked(plain.ID)
	})
	if out == nil || plain.Canceled {
		t.Error("a creature entering as itself was refused")
	}
}

// A permanent spell that can't enter goes to its owner's graveyard (CR
// 608.3e) rather than staying on the stack, and no entry question (riot)
// is asked about an entry that cannot happen.
func TestRefusedPermanentSpellGoesToTheGraveyardUnasked(t *testing.T) {
	g := newWrapGame(t, 2)
	me := g.Seats[g.Turn.ActiveSeat]
	seedEntryBan(t, g, me.ID, EntryRestriction{
		Label:   "Creatures can't enter the battlefield.",
		Forbids: func(q EntryQuery) bool { return q.WouldBe().IsCreature() },
	})
	id := seedKeywordCreature(t, g, "Rioter", KeywordRiot)
	castAndResolve(t, g, id)
	if len(g.PendingChoices) != 0 {
		t.Errorf("an entry question was asked about a refused entry: %v", g.PendingChoices[0].Kind)
	}
	if g.Stack.Contains(id) {
		t.Fatal("the refused permanent spell was left on the stack")
	}
	if enteredBattlefield(g, id) {
		t.Fatal("the refused permanent spell entered")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the refused permanent spell is not in its owner's graveyard")
	}
}

// Nothing is stored: the source leaving lifts the prohibition, and a
// source with no restriction declared costs no look-ahead.
func TestEntryRestrictionStopsWhenItsSourceLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := seedEntryBan(t, g, me.ID, testLandsCantEnter)
	hand := handCardFor(me, "Swamp", "Basic Land — Swamp")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(src) })
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(hand, HandEntryOptions{}); err != nil {
			t.Errorf("hand put: %v", err)
		}
	})
	if !enteredBattlefield(g, hand) {
		t.Error("the prohibition outlived its source")
	}
}
