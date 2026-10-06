package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2115: the revealed-hand pick's variant on the wire. The hand was
// revealed to every seat and a graveyard is public, so every viewer
// gets the same options — the hand, then the graveyard the pick opens —
// the same eligible list, and the bounds ("choose up to one"). Where
// the card goes rides pick_destination.
func TestRevealedHandPickVariantReachesTheWholeTable(t *testing.T) {
	g := buildThreeSeatGame(t)
	caster, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	grave := uuid.New()
	var want []string
	g.WithWriteLock(func() {
		victim.Graveyard.Cards = nil
		victim.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Grave Card", TypeLine: "Instant",
			Owner: victim.ID, Controller: victim.ID,
			KnownBy: map[uuid.UUID]bool{caster.ID: true, victim.ID: true, bystander.ID: true}})
		first := victim.Hand.Cards[0].InstanceID
		want = []string{first.String(), grave.String()}
		if _, err := g.RevealedHandPickForEffect(game.RevealedHandDiscard{
			Chooser: caster.ID, FromPlayer: victim.ID, Count: 1, Reason: "Agonizing Remorse",
			Filter: func(c game.Card) bool { return c.InstanceID == first },
			Label:  "nonland card", Destination: game.PickExile, Optional: true, FromGraveyard: true,
		}); err != nil {
			t.Fatalf("RevealedHandPickForEffect: %v", err)
		}
	})
	handSize := len(victim.Hand.Cards)

	for _, viewer := range []*game.Player{caster, victim, bystander} {
		c := choiceFor(t, ViewOfGameFor(g, viewer.ID.String()), string(game.PendingChoiceRevealedHandPick))
		if len(c.Options) != handSize+1 {
			t.Errorf("%s sees %d options, want the %d revealed cards and the graveyard card", viewer.Name, len(c.Options), handSize)
		}
		for _, o := range c.Options {
			if o.Name == "" || !o.KnownByYou {
				t.Errorf("%s got a back for a revealed or public card: %+v", viewer.Name, o)
			}
		}
		if !slices.Equal(c.Eligible, want) {
			t.Errorf("%s: eligible = %v, want %v", viewer.Name, c.Eligible, want)
		}
		if c.ChooseMin != 0 || c.ChooseMax != 1 || c.PickDestination != "exile" || !c.PickFromGraveyard {
			t.Errorf("%s: min %d max %d destination %q graveyard %v; want 0, 1, exile, true",
				viewer.Name, c.ChooseMin, c.ChooseMax, c.PickDestination, c.PickFromGraveyard)
		}
	}
}

// A mandatory variant's floor is its count, and its destination is
// spelled out even when it is a discard.
func TestRevealedHandPickVariantMandatoryDiscardOnTheWire(t *testing.T) {
	g := buildThreeSeatGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		if _, err := g.RevealedHandPickForEffect(game.RevealedHandDiscard{
			Chooser: caster.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
			Measure: func(game.Card) int { return 1 },
		}); err != nil {
			t.Fatalf("RevealedHandPickForEffect: %v", err)
		}
	})
	c := choiceFor(t, ViewOfGameFor(g, caster.ID.String()), string(game.PendingChoiceRevealedHandPick))
	if c.ChooseMin != 1 || c.ChooseMax != 1 || c.PickDestination != "discard" || c.PickFromGraveyard {
		t.Errorf("min %d max %d destination %q graveyard %v; want 1, 1, discard, false",
			c.ChooseMin, c.ChooseMax, c.PickDestination, c.PickFromGraveyard)
	}
}
