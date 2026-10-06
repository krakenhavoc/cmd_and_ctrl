package game

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// #2115, ADR 0116's 2026-10-05 amendment: the revealed-hand pick's
// variants at the engine level — the destination, the optional floor,
// the graveyard, and the keyed continuation. The cards that use them
// are tested in cards/effects/revealed_hand_variants_test.go.

// rhpSeen records what each test continuation was handed.
type rhpSeen struct {
	calls  int
	chosen []uuid.UUID
	// inHand: for a "first" continuation, whether each chosen card was
	// still in the hand when it ran.
	inHand []bool
	// toughness of the first chosen card, read through the copy.
	toughness int
}

var (
	rhpThenSeen  rhpSeen
	rhpFirstSeen rhpSeen

	rhpThen = RegisterRevealedPickThen(testEffectKeyPrefix+"rhp-then", func(g *Game, r RevealedPick) error {
		rhpThenSeen.calls++
		rhpThenSeen.chosen = r.ChosenIDs()
		if len(r.Chosen) > 0 {
			rhpThenSeen.toughness = r.Chosen[0].Toughness
		}
		return nil
	})
	rhpFirst = RegisterRevealedPickFirst(testEffectKeyPrefix+"rhp-first", func(g *Game, r RevealedPick) error {
		rhpFirstSeen.calls++
		rhpFirstSeen.chosen = r.ChosenIDs()
		from := g.playerByIDLocked(r.FromPlayer)
		for _, id := range r.ChosenIDs() {
			rhpFirstSeen.inHand = append(rhpFirstSeen.inHand, from.Hand.Contains(id))
		}
		return r.Done(g)
	})
)

func rhpReset() {
	rhpThenSeen = rhpSeen{}
	rhpFirstSeen = rhpSeen{}
}

// rhpQueue raises a variant pick and returns it, or fails.
func rhpQueue(t *testing.T, g *Game, d RevealedHandDiscard) *PendingChoice {
	t.Helper()
	var id uuid.UUID
	var err error
	g.WithWriteLock(func() { id, err = g.RevealedHandPickForEffect(d) })
	if err != nil {
		t.Fatalf("RevealedHandPickForEffect: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("no prompt was queued")
	}
	_, c := g.findChoiceLocked(id)
	if c == nil {
		t.Fatal("the queued prompt is not on the queue")
	}
	if c.Kind != PendingChoiceRevealedHandPick {
		t.Fatalf("kind = %q, want %q", c.Kind, PendingChoiceRevealedHandPick)
	}
	return c
}

func rhpEvents(g *Game, kind EventKind, card uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.CardID == card {
			n++
		}
	}
	return n
}

// An exile is not a discard (CR 701.9a): the card goes to exile, no
// EventDiscardCard is emitted for it, and nothing reaches the graveyard.
func TestRevealedHandPickExileIsNotADiscard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	_, spells := rhSetHand(victim, 2)

	c := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
		Filter: rhNonland, Label: "nonland card", Destination: PickExile,
	})
	if c.PickDestination != PickExile {
		t.Fatalf("destination = %q", c.PickDestination)
	}
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, spells[:1]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !g.Exile.Contains(spells[0]) {
		t.Fatal("the chosen card is not in exile")
	}
	if victim.Graveyard.Contains(spells[0]) {
		t.Error("the chosen card reached the graveyard")
	}
	if n := rhpEvents(g, EventDiscardCard, spells[0]); n != 0 {
		t.Errorf("an exile emitted %d discard events", n)
	}
	if victim.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2", victim.Hand.Size())
	}
}

// "You may choose": an empty answer is the decline. Without the
// variant it is refused and the prompt stays open.
func TestRevealedHandPickOptionalTakesTheEmptyAnswer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	rhSetHand(victim, 1)

	mandatory := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
		Filter: rhNonland, Destination: PickExile,
	})
	if err := g.ResolvePendingChoice(mandatory.ID, chooser.ID, nil); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("an empty answer to a mandatory pick: err = %v, want ErrInvalidParam", err)
	}
	if _, c := g.findChoiceLocked(mandatory.ID); c == nil {
		t.Fatal("the refused prompt was dequeued")
	}
	g.WithWriteLock(func() { g.dropChoiceLocked(0) })

	rhpReset()
	optional := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
		Filter: rhNonland, Optional: true, Then: rhpThen,
	})
	if err := g.ResolvePendingChoice(optional.ID, chooser.ID, nil); err != nil {
		t.Fatalf("the decline: %v", err)
	}
	if victim.Hand.Size() != 2 {
		t.Errorf("hand = %d after choosing nothing, want 2", victim.Hand.Size())
	}
	if rhpThenSeen.calls != 1 || len(rhpThenSeen.chosen) != 0 {
		t.Errorf("continuation: %d calls, chosen %v; want one call told nothing", rhpThenSeen.calls, rhpThenSeen.chosen)
	}
}

// The continuation runs after the move and is handed the chosen card
// as it was in the hand.
func TestRevealedHandPickThenRunsAfterTheMoveWithTheChosenCard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	victim.Hand.Cards = nil
	bear := uuid.New()
	victim.Hand.PushTop(Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: victim.ID, Controller: victim.ID})
	rhpReset()

	c := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test", Then: rhpThen,
	})
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !victim.Graveyard.Contains(bear) {
		t.Fatal("the chosen card was not discarded")
	}
	if rhpThenSeen.calls != 1 || !slices.Equal(rhpThenSeen.chosen, []uuid.UUID{bear}) || rhpThenSeen.toughness != 2 {
		t.Errorf("continuation saw %+v, want one call with the Bears and its toughness", rhpThenSeen)
	}
}

// A continuation registered "first" runs while the card is still in the
// hand (CR 608.2c), and the move is its Done.
func TestRevealedHandPickFirstRunsBeforeTheMove(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	_, spells := rhSetHand(victim, 1)
	rhpReset()

	c := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
		Filter: rhNonland, Then: rhpFirst,
	})
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, spells); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !slices.Equal(rhpFirstSeen.inHand, []bool{true}) {
		t.Errorf("the continuation saw the card in the hand: %v, want [true]", rhpFirstSeen.inHand)
	}
	if !victim.Graveyard.Contains(spells[0]) {
		t.Error("Done did not discard the card")
	}
}

// With nothing to choose no prompt goes up, the hand is still revealed,
// and the continuation runs at once, told nothing was chosen — "if you
// don't" happens when there was nothing to choose (CR 609.3).
func TestRevealedHandPickWithNoCandidateRunsTheContinuationAtOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	rhSetHand(victim, 0)
	rhpReset()

	var id uuid.UUID
	var err error
	g.WithWriteLock(func() {
		id, err = g.RevealedHandPickForEffect(RevealedHandDiscard{
			Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
			Filter: rhNonland, Optional: true, Then: rhpThen,
		})
	})
	if err != nil || id != uuid.Nil || len(g.PendingChoices) != 0 {
		t.Fatalf("id %v err %v, %d prompts; want no prompt", id, err, len(g.PendingChoices))
	}
	if rhpThenSeen.calls != 1 || len(rhpThenSeen.chosen) != 0 {
		t.Errorf("continuation: %+v, want one call told nothing", rhpThenSeen)
	}
	for _, seat := range g.Seats {
		if !victim.Hand.Cards[0].IsKnownTo(seat.ID) {
			t.Errorf("%s did not see the revealed land", seat.Name)
		}
	}
}

// "Or a card from their graveyard": every graveyard card is offered,
// unfiltered, after the hand's matches, and can be exiled from there.
// A card in neither place is refused.
func TestRevealedHandPickFromTheGraveyard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	_, spells := rhSetHand(victim, 1)
	victim.Graveyard.Cards = nil
	grave := uuid.New()
	victim.Graveyard.PushTop(Card{InstanceID: grave, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: victim.ID, Controller: victim.ID})

	c := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
		Filter: rhNonland, Destination: PickExile, FromGraveyard: true,
	})
	if want := []uuid.UUID{spells[0], grave}; !slices.Equal(c.DiscardOptions, want) {
		t.Fatalf("options = %v, want the hand's spell then the graveyard land %v", c.DiscardOptions, want)
	}
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{uuid.New()}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a card the prompt never offered: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{grave}); err != nil {
		t.Fatalf("the graveyard card: %v", err)
	}
	if !g.Exile.Contains(grave) || victim.Graveyard.Contains(grave) {
		t.Error("the graveyard card was not exiled")
	}
}

// ADR 0041: the variant pick is plain data. A table waiting on it is a
// restore point, and the restored prompt answers the same way — the
// continuation is found again by its key.
func TestRevealedHandPickIsARestorePoint(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	land, spells := rhSetHand(victim, 1)
	rhpReset()

	rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
		Filter: rhNonland, Optional: true, Destination: PickExile, Then: rhpThen,
	})
	snap := g.CaptureSnapshot()
	if !snap.Continuations.Empty() {
		t.Fatalf("census = %+v, want a restore point", snap.Continuations)
	}
	_, restored := roundTrip(t, g)
	if len(restored.PendingChoices) != 1 {
		t.Fatalf("restored %d prompts", len(restored.PendingChoices))
	}
	c := restored.PendingChoices[0]
	if c.Kind != PendingChoiceRevealedHandPick || !c.PickOptional || c.PickDestination != PickExile || c.PickThen != rhpThen.Key() {
		t.Fatalf("restored prompt = %+v", c)
	}
	rv := restored.playerByIDLocked(victim.ID)
	if err := restored.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{land}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("the restored prompt took the land: %v", err)
	}
	if err := restored.ResolvePendingChoice(c.ID, chooser.ID, spells); err != nil {
		t.Fatalf("the restored prompt's legal pick: %v", err)
	}
	if !restored.Exile.Contains(spells[0]) || rv.Hand.Contains(spells[0]) {
		t.Error("the restored pick did not exile the card")
	}
	if rhpThenSeen.calls != 1 || !slices.Equal(rhpThenSeen.chosen, spells) {
		t.Errorf("continuation after restore: %+v", rhpThenSeen)
	}
}

// A restore point naming a continuation this binary does not register
// was written by a newer one: it is refused, not restored without it.
func TestRevealedHandPickWithAnUnknownKeyIsRefusedAtRestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	rhSetHand(victim, 1)
	rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
		Filter: rhNonland, Then: rhpThen,
	})
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	raw = []byte(strings.Replace(string(raw), rhpThen.Key(), "revealed-pick/from-a-newer-binary", 1))
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, err := decoded.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore: err = %v, want ErrUnknownEffectKey", err)
	}
}

// CR 800.4g: the variant is reassigned like the plain pick.
func TestAReassignedRevealedHandPickVariantKeepsItsShape(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim, controller := g.Seats[0], g.Seats[1], g.Seats[2]
	_, spells := rhSetHand(victim, 1)
	source := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: source, Name: "Source", TypeLine: "Enchantment", Owner: controller.ID, Controller: controller.ID})

	c := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Source: source, Count: 1, Reason: "Test",
		Filter: rhNonland, Optional: true, Destination: PickExile,
	})
	if err := g.Concede(chooser.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	_, moved := g.findChoiceLocked(c.ID)
	if moved == nil || moved.Chooser == chooser.ID {
		t.Fatal("the prompt was not reassigned")
	}
	if err := g.ResolvePendingChoice(moved.ID, moved.Chooser, spells); err != nil {
		t.Fatalf("the new chooser's pick: %v", err)
	}
	if !g.Exile.Contains(spells[0]) {
		t.Error("the reassigned pick did not exile the card")
	}
}
