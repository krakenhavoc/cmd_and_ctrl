package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pending_choice_visibility_test.go covers the one place in
// FilterViewFor where a card from a HIDDEN zone is inlined onto the
// view outside its zone: PendingChoiceView.Options.
//
// viewOfPendingChoices puts the chooser's top library cards on a
// scry / surveil / look-at-top prompt, and the discarder's whole hand
// on a discard_from_hand prompt, so the picker can render faces
// instead of UUIDs. Both pools are hidden zones that the seat
// projection a few lines earlier strips out of every other viewer's
// frame — an opponent's library is replaced wholesale, an opponent's
// unrevealed hand cards are dropped card by card.
//
// Redacting those options instead of dropping them therefore parts
// company with the rest of the file: redactCardForViewer zeroes the
// printed characteristics and KEEPS the instance ID, which is the
// right trade for a face-down permanent (the ID is already on that
// viewer's wire; withholding it from the log would be theatre — S31
// sub-PR 0 settled exactly this) and the wrong one for a card whose
// zone that viewer cannot see at all. A stable UUID for a specific
// card in a hidden zone is a correlation handle: the card is drawn in
// private and cast in public under the same ID, so an opponent who
// wrote down two scry UUIDs learns, three turns later, which of them
// is still up there.
//
// So: an option a viewer is not a knower of is dropped from that
// viewer's copy. The chooser keeps their whole list — they are the
// one picking, and picking a back is a legal answer.

func buildThreeSeatGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 12 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Seat%dCard%02d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 13))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func choiceFor(t *testing.T, v GameView, kind string) PendingChoiceView {
	t.Helper()
	for _, c := range v.PendingChoices {
		if c.Kind == kind {
			return c
		}
	}
	t.Fatalf("no %q choice on the view", kind)
	return PendingChoiceView{}
}

// TestScryOptionsAreNotShippedToOtherSeats is the leak this file
// exists for. A seat scrying two cards must not put two stable
// handles on the top of its library into everybody else's frame.
//
// The COUNT stays public — "scry 2" is a printed number and the
// prompt's Count carries it — which is the same line the zone
// projection takes with an opponent's library size.
func TestScryOptionsAreNotShippedToOtherSeats(t *testing.T) {
	g := buildThreeSeatGame(t)
	scryer, other := g.Seats[0], g.Seats[1]

	var top []uuid.UUID
	var topNames []string
	g.WithWriteLock(func() {
		lib := scryer.Library.Cards
		for i := 0; i < 2; i++ {
			c := &lib[len(lib)-1-i]
			// lookAtTopForEffect marks the chooser — and only the
			// chooser — a knower. That is what makes scry "look at"
			// rather than "reveal".
			c.AddKnower(scryer.ID)
			top = append(top, c.InstanceID)
			topNames = append(topNames, c.Name)
		}
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:       game.PendingChoiceScry,
			Chooser:    scryer.ID,
			FromPlayer: scryer.ID,
			Count:      2,
			Reason:     "Scry 2",
			ScryCards:  top,
		})
	})

	mine := choiceFor(t, ViewOfGameFor(g, scryer.ID.String()), string(game.PendingChoiceScry))
	if len(mine.Options) != 2 {
		t.Fatalf("the scrying seat got %d options, want 2 — it is the one being asked", len(mine.Options))
	}
	for i, o := range mine.Options {
		if o.Name != topNames[i] {
			t.Errorf("the scrying seat's option %d is %q, want %q", i, o.Name, topNames[i])
		}
	}

	theirs := ViewOfGameFor(g, other.ID.String())
	tc := choiceFor(t, theirs, string(game.PendingChoiceScry))
	if len(tc.Options) != 0 {
		t.Errorf("another seat got %d scry options; the top of a library is not theirs to hold handles on: %+v", len(tc.Options), tc.Options)
	}
	if tc.Count != 2 {
		t.Errorf("the scry COUNT should stay public (got %d, want 2) — it is a printed number", tc.Count)
	}
	buf, err := json.Marshal(theirs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for i, id := range top {
		if strings.Contains(string(buf), id.String()) {
			t.Errorf("scryed card %d's instance ID %s reached another seat's frame", i, id)
		}
		if strings.Contains(string(buf), topNames[i]) {
			t.Errorf("scryed card %d's name %q reached another seat's frame", i, topNames[i])
		}
	}
}

// The Thoughtseize shape: the caster revealed the hand and is
// entitled to every card in it; the third seat at the table revealed
// nothing and is entitled to none of them — not the names, and not
// the instance IDs, which the seat projection has already stripped
// out of the hand zone for that same viewer.
func TestDiscardChoiceOptionsFollowTheKnowerSet(t *testing.T) {
	g := buildThreeSeatGame(t)
	caster, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]

	g.WithWriteLock(func() {
		g.QueueDiscardFromRevealedHand(caster.ID, victim.ID, uuid.Nil, 1, "Thoughtseize")
	})
	handSize := len(victim.Hand.Cards)
	if handSize == 0 {
		t.Fatal("the victim has no hand")
	}

	mine := choiceFor(t, ViewOfGameFor(g, caster.ID.String()), string(game.PendingChoiceDiscardFromHand))
	if len(mine.Options) != handSize {
		t.Fatalf("the caster sees %d of the %d cards it revealed", len(mine.Options), handSize)
	}
	for _, o := range mine.Options {
		if o.Name == "" || !o.KnownByYou {
			t.Errorf("the caster got a back for a hand it revealed: %+v", o)
		}
	}

	// The victim knows their own hand and is not the chooser, so the
	// knower rule — not the chooser rule — is what keeps their cards
	// on their own frame.
	theirs := choiceFor(t, ViewOfGameFor(g, victim.ID.String()), string(game.PendingChoiceDiscardFromHand))
	if len(theirs.Options) != handSize {
		t.Errorf("the victim sees %d of their own %d hand cards on the prompt", len(theirs.Options), handSize)
	}

	view := ViewOfGameFor(g, bystander.ID.String())
	bc := choiceFor(t, view, string(game.PendingChoiceDiscardFromHand))
	if len(bc.Options) != 0 {
		t.Errorf("a bystander got %d options out of a hand revealed to somebody else: %+v", len(bc.Options), bc.Options)
	}
	buf, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, c := range victim.Hand.Cards {
		if strings.Contains(string(buf), c.InstanceID.String()) {
			t.Errorf("a bystander's frame carries hand card %q's instance ID %s", c.Name, c.InstanceID)
		}
		if strings.Contains(string(buf), c.Name) {
			t.Errorf("a bystander's frame names hand card %q", c.Name)
		}
	}
}

// The chooser keeps options it does not know. No card produces this
// today — QueueDiscardFromRevealedHand reveals first — but a coercive
// discard that reveals nothing is a legal card to print, and the
// answer to it is "pick one of those backs". Dropping the option for
// its own chooser would leave a prompt that cannot be answered.
func TestChooserKeepsUnknownOptions(t *testing.T) {
	g := buildThreeSeatGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:       game.PendingChoiceDiscardFromHand,
			Chooser:    chooser.ID,
			FromPlayer: victim.ID,
			Count:      1,
			Reason:     "Coercive discard",
		})
	})
	handSize := len(victim.Hand.Cards)

	c := choiceFor(t, ViewOfGameFor(g, chooser.ID.String()), string(game.PendingChoiceDiscardFromHand))
	if len(c.Options) != handSize {
		t.Fatalf("the chooser got %d of %d backs to pick from", len(c.Options), handSize)
	}
	for _, o := range c.Options {
		if o.KnownByYou || o.Name != "" {
			t.Errorf("the chooser was shown a card nobody revealed to it: %+v", o)
		}
		if o.InstanceID == "" {
			t.Error("a back with no instance ID cannot be picked")
		}
	}
}
