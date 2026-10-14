package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// sideboard_test.go — ADR 0191: a deck's sideboard is the player's
// outside-the-game pool, in a container no zone reference reaches.

// sbCard is a sideboard card with a type line.
func sbCard(name, typeLine string) Card {
	c := NewCard(name, uuid.Nil)
	c.TypeLine = typeLine
	c.IsSideboard = true
	return c
}

// newSideboardGame starts a two-seat game in which seat 0 owns `side`
// outside the game, with the mulligans kept.
func newSideboardGame(t *testing.T, side ...Card) *Game {
	t.Helper()
	g := NewGame()
	for i := range 2 {
		deck := buildTestDeck(fmt.Sprintf("Commander %d", i+1))
		if i == 0 {
			deck = append(deck, side...)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func TestSideboardSeatsOutsideTheGame(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"), sbCard("Lightning Bolt", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	side := me.SideboardCards()
	if len(side) != 2 {
		t.Fatalf("sideboard holds %d cards, want 2", len(side))
	}
	if me.Sideboard.Kind != ZoneOutside || me.Sideboard.Owner != me.ID {
		t.Errorf("container is %s owned by %v", me.Sideboard.Kind, me.Sideboard.Owner)
	}
	for _, c := range side {
		if c.Owner != me.ID || c.Controller != me.ID || !c.IsSideboard {
			t.Errorf("%s: owner %v controller %v sideboard %v", c.Name, c.Owner, c.Controller, c.IsSideboard)
		}
		// Its owner may look at it at any time (the Scryfall ruling),
		// so they know it; nobody else does.
		if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
			t.Errorf("%s: known to owner %v, to the opponent %v", c.Name, c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
		}
	}
	if opp.SideboardCards() != nil {
		t.Error("a seat with no sideboard has one")
	}
	// The cards are not part of the deck: 100 cards, none of them these.
	if n := me.Library.Size() + me.Hand.Size() + me.Command.Size(); n != 100 {
		t.Errorf("library + hand + command = %d, want the deck's 100", n)
	}
	// Seating order is the list's order.
	if side[0].Name != "Burning Wish" || side[1].Name != "Lightning Bolt" {
		t.Errorf("sideboard order %q, %q", side[0].Name, side[1].Name)
	}
}

// CR 400.11c: nothing that finds or moves a card by reference can reach
// a card outside the game.
func TestSideboardIsUnreachableByMoveVerbs(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"))
	me := g.Seats[0]
	c := me.SideboardCards()[0]

	g.WithWriteLock(func() {
		if z := g.findCardZoneLocked(c.InstanceID); z != nil {
			t.Errorf("the card index found the sideboard card in %s", z.Kind)
		}
		if _, ok := g.LookupCardForEffect(c.InstanceID); ok {
			t.Error("LookupCardForEffect reached the sideboard card")
		}
		if z := g.zoneFromRefLocked(ZoneRef{Kind: ZoneOutside, Owner: me.ID}); z != nil {
			t.Error("a ZoneRef resolved to the sideboard container")
		}
	})
	for _, to := range []ZoneKind{ZoneHand, ZoneBattlefield, ZoneGraveyard, ZoneExile, ZoneLibrary} {
		owner := me.ID
		if to == ZoneBattlefield || to == ZoneExile {
			owner = uuid.Nil
		}
		for _, from := range []ZoneRef{{Kind: ZoneOutside, Owner: me.ID}, {Kind: ZoneHand, Owner: me.ID}, {Kind: ZoneLibrary, Owner: me.ID}} {
			if err := g.MoveCardByID(from, ZoneRef{Kind: to, Owner: owner}, c.InstanceID); err == nil {
				t.Errorf("the admin move verb moved the sideboard card from %s to %s", from.Kind, to)
			}
		}
	}
	// Effects that act on a card by ID find nothing.
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(c.InstanceID); err == nil {
			t.Error("BounceToHandForEffect reached the sideboard card")
		}
		if err := g.ExileCardForEffect(c.InstanceID); err == nil {
			t.Error("ExileCardForEffect reached the sideboard card")
		}
		if err := g.PutIntoGraveyardForEffect(c.InstanceID); err == nil {
			t.Error("PutIntoGraveyardForEffect reached the sideboard card")
		}
	})
	if got := me.SideboardCards(); len(got) != 1 || got[0].InstanceID != c.InstanceID {
		t.Errorf("the sideboard changed: %d cards", len(got))
	}
}

func TestSideboardCapIsEnforcedAtSeating(t *testing.T) {
	deck := func(n int) []Card {
		d := buildTestDeck("Commander")
		for i := range n {
			d = append(d, sbCard(fmt.Sprintf("Side %d", i), "Instant"))
		}
		return d
	}
	g := NewGame()
	if _, err := g.AddPlayer("P1", deck(MaxSideboardSize+1)); !errors.Is(err, ErrSideboardTooLarge) {
		t.Errorf("AddPlayer with %d sideboard cards: %v, want ErrSideboardTooLarge", MaxSideboardSize+1, err)
	}
	p, err := g.AddPlayer("P1", deck(MaxSideboardSize))
	if err != nil {
		t.Fatalf("AddPlayer with a full sideboard: %v", err)
	}
	if got := len(p.SideboardCards()); got != MaxSideboardSize {
		t.Errorf("sideboard holds %d, want %d", got, MaxSideboardSize)
	}
	if err := g.ReplaceDeck(p.ID, deck(MaxSideboardSize+1)); !errors.Is(err, ErrSideboardTooLarge) {
		t.Errorf("ReplaceDeck: %v, want ErrSideboardTooLarge", err)
	}
	if got := len(p.SideboardCards()); got != MaxSideboardSize {
		t.Errorf("a refused ReplaceDeck changed the sideboard to %d cards", got)
	}
}

func TestReplaceDeckReplacesTheSideboard(t *testing.T) {
	g := NewGame()
	p, err := g.AddPlayer("P1", append(buildTestDeck("Commander"), sbCard("Old", "Sorcery")))
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ReplaceDeck(p.ID, buildTestDeck("Other")); err != nil {
		t.Fatal(err)
	}
	if len(p.SideboardCards()) != 0 {
		t.Error("a deck without a sideboard kept the old one")
	}
	if err := g.ReplaceDeck(p.ID, append(buildTestDeck("Third"), sbCard("New", "Instant"), sbCard("Newer", "Instant"))); err != nil {
		t.Fatal(err)
	}
	if got := p.SideboardCards(); len(got) != 2 || got[0].Name != "New" {
		t.Errorf("sideboard after the third deck: %+v", got)
	}
}

func TestSideboardLeavesWithItsOwner(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"))
	me := g.Seats[0]
	g.WithWriteLock(func() { g.removeObjectsOwnedByLocked(me.ID) })
	if len(me.SideboardCards()) != 0 {
		t.Error("a sideboard stayed after its owner left")
	}
}

func TestSideboardAndCompanionAreSeparateContainers(t *testing.T) {
	g := NewGame()
	deck := append(buildTestDeck("Commander"), testCompanion(), sbCard("Burning Wish", "Sorcery"))
	p, err := g.AddPlayer("P1", deck)
	if err != nil {
		t.Fatal(err)
	}
	if c, ok := p.CompanionCard(); !ok || c.IsSideboard {
		t.Errorf("companion %+v ok %v", c, ok)
	}
	if got := p.SideboardCards(); len(got) != 1 || got[0].Name != "Burning Wish" {
		t.Errorf("sideboard %+v", got)
	}
	// The companion special action is offered for the companion only.
	if CompanionOffer(p.SideboardCards()[0]) != nil {
		t.Error("a sideboard card offers the companion action")
	}
}

func TestSideboardSurvivesCloneUndoAndSnapshot(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"), sbCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]
	wish := me.SideboardCards()[0]

	// Clone: an independent copy.
	clone := g.Clone()
	if got := clone.Seats[0].SideboardCards(); len(got) != 2 || got[0].InstanceID != wish.InstanceID {
		t.Fatalf("Clone lost the sideboard: %+v", got)
	}

	// Undo: bring a card in, then restore the clone taken before.
	g.WithWriteLock(func() {
		if err := g.BringFromOutsideForEffect(me.ID, wish.InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	if len(me.SideboardCards()) != 1 {
		t.Fatalf("the sideboard did not shrink: %d", len(me.SideboardCards()))
	}
	g.RestoreFrom(clone)
	if got := g.Seats[0].SideboardCards(); len(got) != 2 {
		t.Errorf("undo left %d sideboard cards, want 2", len(got))
	}
	for _, c := range g.Seats[0].Hand.Cards {
		if c.InstanceID == wish.InstanceID {
			t.Error("undo left the brought-in card in the hand")
		}
	}

	// Snapshot: capture, JSON, restore.
	_, restored := roundTrip(t, g)
	rs := restored.Seats[0].SideboardCards()
	if len(rs) != 2 || rs[0].InstanceID != wish.InstanceID || !rs[0].IsSideboard || rs[0].Name != "Burning Wish" {
		t.Fatalf("restored sideboard %+v", rs)
	}
	if restored.Seats[0].Sideboard.Kind != ZoneOutside || restored.Seats[0].Sideboard.Owner != restored.Seats[0].ID {
		t.Errorf("restored container %s owned by %v", restored.Seats[0].Sideboard.Kind, restored.Seats[0].Sideboard.Owner)
	}
	if !rs[0].IsKnownTo(restored.Seats[0].ID) {
		t.Error("the owner lost sight of their own sideboard in the round trip")
	}
	if restored.Seats[1].SideboardCards() != nil {
		t.Error("a seat with no sideboard has one after the round trip")
	}
	// And it can still be used.
	restored.WithWriteLock(func() {
		if err := restored.BringFromOutsideForEffect(restored.Seats[0].ID, wish.InstanceID); err != nil {
			t.Errorf("bring after restore: %v", err)
		}
	})
}

// ADR 0191 §4: a file written before a sideboard could be seated has no
// key and restores with none; a binary that does not know the key drops
// it, and the player loses what an effect could have brought in.
func TestRestoreWithoutTheSideboardKeyLosesOnlyTheOption(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"))
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	seat := doc["seats"].([]any)[0].(map[string]any)
	if _, ok := seat["sideboard"]; !ok {
		t.Fatal("the snapshot has no sideboard key to drop")
	}
	// A seat with none writes no key at all.
	if _, ok := doc["seats"].([]any)[1].(map[string]any)["sideboard"]; ok {
		t.Error("a seat with no sideboard wrote the key")
	}
	delete(seat, "sideboard")
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restored.Seats[0].SideboardCards() != nil {
		t.Error("a file without the key restored a sideboard")
	}
	if n := restored.Seats[0].Library.Size() + restored.Seats[0].Hand.Size() + restored.Seats[0].Command.Size(); n != 100 {
		t.Errorf("deck size after the restore = %d, want 100", n)
	}
}

func TestOutsideCandidatesReadTheSideboardOnlyForItsOwner(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"), sbCard("Lightning Bolt", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if got := g.OutsideCandidatesForEffect(me.ID, nil); len(got) != 2 {
			t.Errorf("owner's candidates = %d, want 2", len(got))
		}
		if got := g.OutsideCandidatesForEffect(me.ID, func(c Card) bool { return c.IsSorcery() }); len(got) != 1 {
			t.Errorf("sorcery candidates = %d, want 1", len(got))
		}
		if got := g.OutsideCandidatesForEffect(opp.ID, nil); got != nil {
			t.Errorf("an opponent's candidates came from my sideboard: %v", got)
		}
	})
}

// ADR 0167 §5 (CR 723.4): while another player makes the owner's
// decisions, nothing is offered from the owner's sideboard.
func TestOutsideCandidatesAreClosedToAControlledPlayer(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"))
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.PlayerControls = append(g.PlayerControls, PlayerControl{
			Controller: opp.ID, Controlled: me.ID, Until: ControlUntilTurn, Stage: ControlTurn, TurnSeq: g.Turn.Seq,
		})
		if g.ControllerOfPlayerLocked(me.ID) != opp.ID {
			t.Fatal("the control record is not live in this fixture")
		}
		if got := g.OutsideCandidatesForEffect(me.ID, nil); got != nil {
			t.Errorf("a controlled player's sideboard offered %v", got)
		}
		if g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID}) {
			t.Error("a prompt went up over a closed sideboard")
		}
	})
}

func TestBringFromOutsidePutsTheCardInTheHandAsAnEvent(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"), sbCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]
	bolt := me.SideboardCards()[1]
	before := len(g.Events)
	handBefore := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.BringFromOutsideForEffect(me.ID, bolt.InstanceID); err != nil {
			t.Fatal(err)
		}
		if err := g.BringFromOutsideForEffect(me.ID, bolt.InstanceID); err == nil {
			t.Error("the same card came in twice")
		}
		if err := g.BringFromOutsideForEffect(g.Seats[1].ID, bolt.InstanceID); err == nil {
			t.Error("an opponent brought in my card")
		}
	})
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), handBefore+1)
	}
	if len(me.SideboardCards()) != 1 {
		t.Errorf("sideboard %d, want 1", len(me.SideboardCards()))
	}
	moved := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventZoneMove && ev.CardID == bolt.InstanceID && ev.OldZone == ZoneOutside && ev.NewZone == ZoneHand {
			moved = true
		}
	}
	if !moved {
		t.Error("no zone-move event from outside the game to the hand")
	}
}

// RevealOutsideForEffect makes every seat a knower and announces it; the
// card stays where it is.
func TestRevealOutsideMakesTheTableKnowTheCard(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"), sbCard("Lightning Bolt", "Instant"))
	me, opp := g.Seats[0], g.Seats[1]
	wish := me.SideboardCards()[0]
	g.WithWriteLock(func() {
		seq := g.RevealOutsideForEffect(RevealSpec{Player: me.ID, Reason: "test", Cards: []uuid.UUID{wish.InstanceID, uuid.New()}})
		if seq == 0 {
			t.Fatal("nothing was revealed")
		}
	})
	got := me.SideboardCards()
	if !got[0].IsKnownTo(opp.ID) {
		t.Error("the opponent did not learn the revealed card")
	}
	if got[1].IsKnownTo(opp.ID) {
		t.Error("the card that was not revealed became known")
	}
	if len(got) != 2 {
		t.Error("a reveal moved a card")
	}
	reveals := 0
	for _, ev := range g.Events {
		if ev.Kind == EventRevealCards && ev.CardID == wish.InstanceID && ev.OldZone == ZoneOutside {
			reveals++
		}
	}
	if reveals != 1 {
		t.Errorf("%d reveal events, want 1", reveals)
	}
}

func TestQueueOutsidePickAsksOnlyAboutMatchingCards(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"), sbCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Question: "a land", Match: func(c Card) bool { return c.IsLand() },
		}) {
			t.Error("a prompt went up with nothing to choose from")
		}
		if len(g.PendingChoices) != 0 {
			t.Errorf("%d choices queued", len(g.PendingChoices))
		}
		if !g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Question: "an instant", Match: func(c Card) bool { return c.IsInstant() },
		}) {
			t.Fatal("no prompt for an instant")
		}
		c := g.PendingChoices[0]
		if c.Kind != PendingChoiceChooseCards || c.Chooser != me.ID || len(c.ChooseCards) != 1 || c.ChooseMax != 1 || c.ChooseMin != 0 {
			t.Errorf("prompt %+v", c)
		}
		if c.ChooseDestination != ChooseFromOutside {
			t.Errorf("destination %q", c.ChooseDestination)
		}
	})
}

func TestOutsidePickValidatesTheSet(t *testing.T) {
	g := newSideboardGame(t,
		sbCard("Alpha", "Sorcery"), sbCard("Alpha", "Sorcery"), sbCard("Beta", "Sorcery"))
	me := g.Seats[0]
	side := me.SideboardCards()
	var got []uuid.UUID
	var choiceID uuid.UUID
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Question: "two", Min: 0, Max: 2,
			Validate: func(picked []Card) bool { return len(picked) == 2 && picked[0].Name != picked[1].Name },
			Then: func(_ *Game, picked []uuid.UUID) error {
				got = picked
				return nil
			},
		})
		choiceID = g.PendingChoices[0].ID
	})
	// Two cards with the same name: refused, prompt stays open.
	if err := g.ResolveChooseCards(choiceID, me.ID, []uuid.UUID{side[0].InstanceID, side[1].InstanceID}); !errors.Is(err, ErrChoiceSetRejected) {
		t.Errorf("same-name pair: %v, want ErrChoiceSetRejected", err)
	}
	// One card: refused.
	if err := g.ResolveChooseCards(choiceID, me.ID, []uuid.UUID{side[0].InstanceID}); !errors.Is(err, ErrChoiceSetRejected) {
		t.Errorf("single card: %v, want ErrChoiceSetRejected", err)
	}
	// A card that is not a candidate at all.
	if err := g.ResolveChooseCards(choiceID, me.ID, []uuid.UUID{uuid.New(), side[0].InstanceID}); err == nil {
		t.Error("an unknown card was accepted")
	}
	// Somebody else's answer.
	if err := g.ResolveChooseCards(choiceID, g.Seats[1].ID, []uuid.UUID{side[0].InstanceID, side[2].InstanceID}); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("opponent's answer: %v, want ErrNotTheChooser", err)
	}
	if len(got) != 0 {
		t.Fatal("Then ran for a refused answer")
	}
	// Two different names: accepted.
	if err := g.ResolveChooseCards(choiceID, me.ID, []uuid.UUID{side[0].InstanceID, side[2].InstanceID}); err != nil {
		t.Fatalf("valid pair: %v", err)
	}
	if len(got) != 2 || got[0] != side[0].InstanceID || got[1] != side[2].InstanceID {
		t.Errorf("Then saw %v", got)
	}
}

// A card that leaves the sideboard under an open prompt is pruned from
// it, and the zone re-check refuses it if an answer still names it.
func TestOutsidePickPrunesACardThatLeft(t *testing.T) {
	g := newSideboardGame(t, sbCard("Alpha", "Sorcery"), sbCard("Beta", "Sorcery"), sbCard("Gamma", "Sorcery"))
	me := g.Seats[0]
	side := me.SideboardCards()
	var choiceID uuid.UUID
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{Player: me.ID, Question: "one", Max: 3})
		choiceID = g.PendingChoices[0].ID
		if err := g.BringFromOutsideForEffect(me.ID, side[2].InstanceID); err != nil {
			t.Fatal(err)
		}
		if g.pickStillOutsideLocked(g.PendingChoices[0], side[2].InstanceID) {
			t.Error("the zone re-check still finds a card that left")
		}
		g.pruneCardSetChoicesLocked()
		if len(g.PendingChoices) != 1 || len(g.PendingChoices[0].ChooseCards) != 2 {
			t.Fatalf("pruned prompt %+v", g.PendingChoices)
		}
	})
	if err := g.ResolveChooseCards(choiceID, me.ID, []uuid.UUID{side[2].InstanceID}); err == nil {
		t.Error("an answer naming a card that left was accepted")
	}
	if err := g.ResolveChooseCards(choiceID, me.ID, []uuid.UUID{side[0].InstanceID, side[1].InstanceID}); err != nil {
		t.Errorf("answer naming the two that remain: %v", err)
	}
}

func TestEmptyPickOfAnOutsidePromptRunsThenWithNothing(t *testing.T) {
	g := newSideboardGame(t, sbCard("Burning Wish", "Sorcery"))
	me := g.Seats[0]
	ran := false
	var choiceID uuid.UUID
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(OutsidePickPrompt{
			Player: me.ID, Question: "x",
			Then: func(_ *Game, picked []uuid.UUID) error {
				ran = len(picked) == 0
				return nil
			},
		})
		choiceID = g.PendingChoices[0].ID
	})
	if err := g.ResolveChooseCards(choiceID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Error("the continuation did not run for an empty pick")
	}
	if len(me.SideboardCards()) != 1 {
		t.Error("declining took the card")
	}
}
