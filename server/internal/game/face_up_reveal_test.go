package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// face_up_reveal_test.go — #6249: "Morph—Reveal a green card in your
// hand" (Temur Charger). CR 702.37e: turning the permanent face up is a
// special action (CR 116.2b) whose cost is paid first; here the cost is
// revealing a green card from the controller's hand (CR 701.20a), which
// stays there (CR 701.20b).

// revealMorphGame is morphGame with a "Morph—Reveal a green card"
// declaration, cast face down and resolved.
func revealMorphGame(t *testing.T) (*Game, *Player, uuid.UUID) {
	t.Helper()
	offer := morphOffer(FaceDownMorphed, "", false)
	offer.FaceDown.FaceUpReveal = "G"
	g, me, id := morphGame(t, offer)
	castAndLandFaceDown(t, g, me, id, "morph")
	return g, me, id
}

func coloredCard(name string, owner uuid.UUID, color string) Card {
	c := NewCard(name, owner)
	c.TypeLine = "Creature — Bear"
	c.Colors = []string{color}
	return c
}

func turnUpOffers(g *Game, me *Player, id uuid.UUID) []SpecialAction {
	var out []SpecialAction
	c, _, _ := cardAnywhere(g, id)
	g.ReadSnapshot(func() {
		for _, sa := range g.SpecialActionsOfferedLocked(me.ID, c, ZoneBattlefield) {
			if sa.Kind == SpecialActionTurnFaceUp {
				out = append(out, sa)
			}
		}
	})
	return out
}

// TestRevealMorphTurnsFaceUpByRevealingAGreenCard is the acceptance:
// naming a green card in hand turns the permanent up, spends no mana,
// and reveals that card to the table without moving it.
func TestRevealMorphTurnsFaceUpByRevealingAGreenCard(t *testing.T) {
	g, me, id := revealMorphGame(t)
	green := coloredCard("Grizzly Bears", me.ID, "G")
	g.WithWriteLock(func() { me.Hand.PushTop(green) })

	offers := turnUpOffers(g, me, id)
	if len(offers) != 1 || offers[0].RevealColor != "G" || offers[0].Cost != "" {
		t.Fatalf("offers = %+v, want one reveal-green offer with no mana cost", offers)
	}
	if offers[0].Label != "Turn face up — reveal a green card" {
		t.Errorf("label = %q", offers[0].Label)
	}
	before := me.ManaPool.Total()
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, RevealID: green.InstanceID}); err != nil {
		t.Fatalf("turn face up revealing a green card: %v", err)
	}
	if c, _, _ := cardAnywhere(g, id); c.FaceDown {
		t.Fatal("still face down after the special action")
	}
	if got := me.ManaPool.Total(); got != before {
		t.Errorf("mana %d → %d: a reveal cost spends none", before, got)
	}
	if !me.Hand.Contains(green.InstanceID) {
		t.Error("the revealed card left the hand (CR 701.20b)")
	}
	var revealed bool
	for _, ev := range g.Events {
		if ev.Kind == EventRevealCards && ev.CardID == green.InstanceID {
			revealed = true
		}
	}
	if !revealed {
		t.Error("the green card was not revealed (CR 701.20a)")
	}
}

// TestRevealMorphIsNotOfferedWithoutAGreenCard: a hand holding no green
// card cannot pay the cost, so the special action is not offered, and
// taking it anyway is refused.
func TestRevealMorphIsNotOfferedWithoutAGreenCard(t *testing.T) {
	g, me, id := revealMorphGame(t)
	red := coloredCard("Raging Goblin", me.ID, "R")
	g.WithWriteLock(func() { me.Hand.PushTop(red) })
	if offers := turnUpOffers(g, me, id); len(offers) != 0 {
		t.Fatalf("offers = %+v with no green card in hand, want none", offers)
	}
	err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{RevealID: red.InstanceID})
	if !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Fatalf("turn face up revealing a red card = %v, want ErrSpecialActionNotOffered", err)
	}
	if c, _, _ := cardAnywhere(g, id); !c.FaceDownIsPermanent() {
		t.Fatal("the permanent turned up without its cost")
	}
}

// TestRevealMorphRefusesAWrongOrMissingReveal: with a green card in
// hand the offer stands, but the payload must name a green card in the
// actor's hand. Naming none, a red one, or one that is not in hand is
// refused before anything happens; a reveal named for a mana morph is
// refused too.
func TestRevealMorphRefusesAWrongOrMissingReveal(t *testing.T) {
	g, me, id := revealMorphGame(t)
	green := coloredCard("Grizzly Bears", me.ID, "G")
	red := coloredCard("Raging Goblin", me.ID, "R")
	g.WithWriteLock(func() {
		me.Hand.PushTop(green)
		me.Hand.PushTop(red)
	})
	for name, rid := range map[string]uuid.UUID{
		"none":        uuid.Nil,
		"a red card":  red.InstanceID,
		"not in hand": uuid.New(),
	} {
		err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{RevealID: rid})
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("revealing %s = %v, want ErrInvalidParam", name, err)
		}
	}
	if c, _, _ := cardAnywhere(g, id); !c.FaceDownIsPermanent() {
		t.Fatal("a refused action turned the permanent up")
	}

	g2, me2, id2 := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	castAndLandFaceDown(t, g2, me2, id2, "morph")
	g2.WithWriteLock(func() { me2.Hand.PushTop(coloredCard("Grizzly Bears", me2.ID, "G")) })
	var rid uuid.UUID
	g2.ReadSnapshot(func() { rid = me2.Hand.Cards[0].InstanceID })
	err := g2.PerformSpecialAction(me2.ID, id2, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true, RevealID: rid})
	if !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a reveal on morph {1}{U} = %v, want ErrInvalidParam", err)
	}
}

// TestFaceUpRevealOptionsListTheMatchingHandCards is the view's list:
// the green cards in the controller's hand, and nothing for an offer
// with no reveal cost.
func TestFaceUpRevealOptionsListTheMatchingHandCards(t *testing.T) {
	g, me, id := revealMorphGame(t)
	g1 := coloredCard("Grizzly Bears", me.ID, "G")
	r := coloredCard("Raging Goblin", me.ID, "R")
	g2 := coloredCard("Llanowar Elves", me.ID, "G")
	g.WithWriteLock(func() {
		me.Hand.PushTop(g1)
		me.Hand.PushTop(r)
		me.Hand.PushTop(g2)
	})
	offers := turnUpOffers(g, me, id)
	if len(offers) != 1 {
		t.Fatalf("offers = %+v, want one", offers)
	}
	var got []uuid.UUID
	g.ReadSnapshot(func() { got = g.FaceUpRevealOptionsForEffect(me.ID, offers[0]) })
	want := map[uuid.UUID]bool{g1.InstanceID: true, g2.InstanceID: true}
	if len(got) != 2 || !want[got[0]] || !want[got[1]] {
		t.Errorf("options = %v, want the two green cards", got)
	}
	g.ReadSnapshot(func() {
		if opts := g.FaceUpRevealOptionsForEffect(me.ID, SpecialAction{Kind: SpecialActionTurnFaceUp, Cost: "{1}{U}"}); opts != nil {
			t.Errorf("options for a mana morph = %v, want none", opts)
		}
	})
}
