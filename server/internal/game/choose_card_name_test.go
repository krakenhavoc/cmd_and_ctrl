package game

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// choose_card_name_test.go — #1210. The as-enters family's fourth
// member, and the three things that make it different from the three
// before it: there is no vocabulary to validate against, the answer
// is compared by CardNameMatches rather than by ==, and it must
// survive a clone and a snapshot without becoming copiable.

// queueName puts a permanent on the battlefield and asks its
// controller to name a card, returning the permanent and the choice.
func queueName(t *testing.T, g *Game, owner *Player) (uuid.UUID, uuid.UUID) {
	t.Helper()
	src := pushGateCard(g, "Pithing Needle", "Artifact", needleOracle, owner.ID)
	var choice uuid.UUID
	g.WithWriteLock(func() {
		choice = g.QueueCardNameChoiceForEffect(owner.ID, src, "Pithing Needle — choose a card name")
	})
	if choice == uuid.Nil {
		t.Fatal("QueueCardNameChoiceForEffect queued nothing")
	}
	return src, choice
}

// TestAChosenCardNameIsFreeTextCheckedOnlyForShape is the decision
// that separates this prompt from its three siblings: CR 201.2 lets a
// player name ANY card name, so there is nothing to validate against
// and the engine checks the shape alone.
func TestAChosenCardNameIsFreeTextCheckedOnlyForShape(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, choice := queueName(t, g, me)

	// A name this server has never heard of is a legal answer.
	if err := g.ResolveCardNameChoice(choice, me.ID, "  Some Unreleased Card  "); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	var got string
	g.ReadSnapshot(func() { got = g.ChosenNameOf(src) })
	if got != "Some Unreleased Card" {
		t.Errorf("ChosenName = %q, want the trimmed free text", got)
	}

	// Shape, though, is checked: empty and over-long are refused, and
	// a refusal leaves the prompt outstanding rather than eating it.
	_, choice2 := queueName(t, g, me)
	if err := g.ResolveCardNameChoice(choice2, me.ID, "   "); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("empty name: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveCardNameChoice(choice2, me.ID, strings.Repeat("x", MaxChosenNameLen+1)); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("over-long name: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveCardNameChoice(choice2, me.ID, "Sol Ring"); err != nil {
		t.Fatalf("the prompt should still be open after two refusals: %v", err)
	}
}

// TestCardNameMatchesAsksEveryFace is CR 201.2b. Card.Name is the
// ACTIVE face's name, so a comparison that used it alone would stop
// matching the moment a named permanent transformed — and naming one
// half of a split card names the card.
func TestCardNameMatchesAsksEveryFace(t *testing.T) {
	c := Card{
		Name: "Moonrage Brute",
		Faces: []Face{
			{Name: "Brutal Cathar"},
			{Name: "Moonrage Brute"},
		},
		ActiveFace: 1,
	}
	for _, want := range []string{"Moonrage Brute", "Brutal Cathar", "brutal cathar", "  Brutal Cathar "} {
		if !CardNameMatches(c, want) {
			t.Errorf("CardNameMatches(%q) = false, want true", want)
		}
	}
	for _, no := range []string{"", "   ", "Sol Ring"} {
		if CardNameMatches(c, no) {
			t.Errorf("CardNameMatches(%q) = true, want false", no)
		}
	}
}

// TestAChosenNameSurvivesACloneAndIsClearedOnTheWayOut pins both ends
// of the lifecycle: the undo snapshot carries the answer (nothing can
// re-derive a player's free text), and CR 400.7 clears it, so a
// bounced Needle names again and one in a graveyard restricts nobody.
func TestAChosenNameSurvivesACloneAndIsClearedOnTheWayOut(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, choice := queueName(t, g, me)
	if err := g.ResolveCardNameChoice(choice, me.ID, "Sol Ring"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}

	clone := g.Clone()
	var cloned string
	clone.ReadSnapshot(func() { cloned = clone.ChosenNameOf(src) })
	if cloned != "Sol Ring" {
		t.Errorf("after Clone: ChosenName = %q, want %q", cloned, "Sol Ring")
	}

	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	g.ReadSnapshot(func() {
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == src && c.ChosenName != "" {
				t.Errorf("in the graveyard: ChosenName = %q, want empty (CR 400.7)", c.ChosenName)
			}
		}
	})
}

// TestAChosenNameIsNotCopiable is CR 706.2, and it falls out of where
// the field lives rather than out of a rule anybody wrote: the
// copiable values are the printed characteristics, and this is not
// one of them. A Clone of a Pithing Needle names its own card.
func TestAChosenNameIsNotCopiable(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, choice := queueName(t, g, me)
	if err := g.ResolveCardNameChoice(choice, me.ID, "Sol Ring"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	var vals PrintedValues
	g.ReadSnapshot(func() {
		live, _ := g.LookupCardForEffect(src)
		vals = CopiableValuesOf(live)
	})
	// The projection has no slot for it at all — which is the
	// assertion. If a field is ever added, this test is where the
	// rules question gets asked again.
	if strings.Contains(fmt.Sprintf("%+v", vals), "Sol Ring") {
		t.Error("the chosen name reached the copiable values — CR 706.2 says a copy chooses its own")
	}
}
