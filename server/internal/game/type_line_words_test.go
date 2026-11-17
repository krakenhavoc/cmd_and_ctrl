package game

import (
	"testing"

	"github.com/google/uuid"
)

// type_line_words_test.go — #5213. typeLineHas matches whole words of
// the type line, never a substring: a Lander token ("Token Artifact —
// Lander") is not a land, so its entering is not a land entering.

func TestTypeLineHasMatchesWholeWordsOnly(t *testing.T) {
	for _, tc := range []struct {
		line, needle string
		want         bool
	}{
		// The bug: Lander is a subtype, and it is not the card type Land.
		{"Token Artifact — Lander", "land", false},
		{"Token Artifact — Lander", "lander", true},
		{"Token Artifact — Lander", "artifact", true},
		{"Token Artifact — Lander", "token", true},
		{"Token Artifact — Lander", "art", false},
		// Real lands still answer "land".
		{"Basic Land — Forest", "land", true},
		{"Basic Land — Forest", "basic", true},
		{"Basic Land — Forest", "forest", true},
		{"Token Land — Forest", "land", true},
		{"Land", "land", true},
		{"Artifact Land", "land", true},
		{"Land Creature — Forest Dryad", "creature", true},
		// Card types, supertypes and subtypes, any capitalisation.
		{"Legendary Creature — Human Wizard", "legendary", true},
		{"Legendary Creature — Human Wizard", "creature", true},
		{"Legendary Creature — Human Wizard", "Creature", true},
		{"Legendary Creature — Human Wizard", "wizard", true},
		{"Legendary Creature — Human Wizard", "man", false},
		{"Legendary Enchantment Creature — God", "enchantment", true},
		// A subtype that contains a card type's letters is not that type.
		{"Creature — Elemental Island", "land", false},
		{"Enchantment — Cartouche", "art", false},
		{"Legendary Planeswalker — Ugin", "planeswalker", true},
		// A hyphenated subtype is one word.
		{"Artifact Creature — Assembly-Worker", "assembly-worker", true},
		{"Artifact Creature — Assembly-Worker", "worker", false},
		// Both faces of a split or modal line are read.
		{"Instant // Sorcery", "instant", true},
		{"Instant // Sorcery", "sorcery", true},
		{"Creature — Human//Land", "land", true},
		// No dash, an ASCII stand-in, or extra spaces.
		{"Creature—Bear", "creature", true},
		{"Creature—Bear", "bear", true},
		{"Creature - Bear", "bear", true},
		{"  Artifact  ", "artifact", true},
		// Empty inputs and multi-word needles match nothing.
		{"", "land", false},
		{"Land", "", false},
		{"Basic Land — Forest", "basic land", false},
	} {
		if got := typeLineHas(tc.line, tc.needle); got != tc.want {
			t.Errorf("typeLineHas(%q, %q) = %v, want %v", tc.line, tc.needle, got, tc.want)
		}
	}
}

func TestALanderTokenIsNotALand(t *testing.T) {
	lander := Card{Name: "Lander", TypeLine: "Token Artifact — Lander"}
	if lander.PrintedIsLand() || lander.HasCardType("land") || lander.IsLand() {
		t.Error("a Lander token reads as a land")
	}
	if !lander.HasCardType("artifact") || !lander.HasSubtype("Lander") || !IsTokenPtr(&lander) {
		t.Error("a Lander token lost its artifact type, its Lander subtype or its token status")
	}
}

func TestTurnTallyDoesNotCountALanderEnteringAsALand(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	lander, forest := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: lander, Name: "Lander", TypeLine: "Token Artifact — Lander", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
	})
	emit(g, Event{Kind: EventETB, CardID: lander})
	if got := g.TurnTallyFor(me).LandsEntered; got != 0 {
		t.Fatalf("a Lander entering counted as %d lands, want 0", got)
	}
	emit(g, Event{Kind: EventETB, CardID: forest})
	if got := g.TurnTallyFor(me).LandsEntered; got != 1 {
		t.Errorf("lands entered after a Forest: %d, want 1", got)
	}
}

func TestCreatedLanderTokenDoesNotCountAsALandEntering(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me, Card{Name: "Lander", TypeLine: "Token Artifact — Lander"}, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := g.TurnTallyFor(me).LandsEntered; got != 0 {
		t.Fatalf("two Lander tokens counted as %d lands entering, want 0", got)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me, Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := g.TurnTallyFor(me).LandsEntered; got != 1 {
		t.Errorf("a land token entering: %d lands entered, want 1", got)
	}
}
