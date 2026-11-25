package game

import (
	"testing"

	"github.com/google/uuid"
)

// card_copy_cost_less_test.go — #6231: "copy that card and you may cast
// the copy. That copy costs {2} less to cast" (God-Eternal Kefnet). The
// reduction is a CR 601.2f one: "Some effects may increase or reduce the
// cost to pay". CR 707.12: the copy is created in the same zone the
// object is in, so a card revealed from the hand is copied there.

// kefnetOffer is the offer Kefnet's trigger makes.
func kefnetOffer(p *Player) CardCopyOffer {
	return CardCopyOffer{Player: p.ID, SourceName: "Test Kefnet", Timing: TimingFlash, Lapse: LapseStaysInExile, CostLess: 2, InHand: true}
}

func copyFixtureInHand(p *Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// A {3}{U} instant drawn and revealed: the copy is made in the hand and
// costs {1}{U}, so two mana casts it.
func TestCardCopyCostLessReducesTheCopysPrice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	orig := copyFixtureInHand(me, "Fixture Insight", "Instant", "{3}{U}")
	copies := resolveCopyingAbility(t, g, me, kefnetOffer(me), orig)
	if len(copies) != 1 {
		t.Fatalf("copies = %v, want one", copies)
	}
	if !me.Hand.Contains(copies[0]) {
		t.Fatal("the copy of a card in hand is not in the hand (CR 707.12)")
	}
	c := resolutionCastPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no resolution cast prompt over the copy")
	}
	me.ManaPool.AddMana(ManaToken{Color: "U"}, ManaToken{Color: "C"})
	if err := g.CastForResolutionChoice(me.ID, c.ID, copies[0], CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Fatalf("cast the copy for {1}{U}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left after the cast, want 0", me.ManaPool.Total())
	}
	if !me.Hand.Contains(orig) {
		t.Error("the original left the hand")
	}
}

// The reduction is the copy's only: with one mana short of {1}{U} the
// copy can't be cast.
func TestCardCopyCostLessIsNotFree(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	orig := copyFixtureInHand(me, "Fixture Insight", "Instant", "{3}{U}")
	copies := resolveCopyingAbility(t, g, me, kefnetOffer(me), orig)
	c := resolutionCastPrompt(g, me.ID)
	if c == nil || len(copies) != 1 {
		t.Fatal("no prompt over the copy")
	}
	me.ManaPool.AddMana(ManaToken{Color: "U"})
	if err := g.CastForResolutionChoice(me.ID, c.ID, copies[0], CastSpellParams{FromZone: "hand", Strict: true}); err == nil {
		t.Fatal("a {3}{U} copy two less was cast for {U}")
	}
}

// A graveyard copy takes the reduction too.
func TestCardCopyCostLessInTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	orig := copyFixtureInGraveyard(me, "Fixture Insight", "Instant", "{3}{U}")
	offer := kefnetOffer(me)
	offer.InHand = false
	copies := resolveCopyingAbility(t, g, me, offer, orig)
	c := resolutionCastPrompt(g, me.ID)
	if c == nil || len(copies) != 1 {
		t.Fatal("no prompt over the copy")
	}
	me.ManaPool.AddMana(ManaToken{Color: "U"}, ManaToken{Color: "C"})
	if err := g.CastForResolutionChoice(me.ID, c.ID, copies[0], CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the graveyard copy for {1}{U}: %v", err)
	}
}

// Declined, the hand copy ceases to exist (CR 704.5e): it is not left in
// the hand to be cast later at its printed cost.
func TestCardCopyInHandCeasesWhenDeclined(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	orig := copyFixtureInHand(me, "Fixture Insight", "Instant", "{3}{U}")
	copies := resolveCopyingAbility(t, g, me, kefnetOffer(me), orig)
	c := resolutionCastPrompt(g, me.ID)
	if c == nil || len(copies) != 1 {
		t.Fatal("no prompt over the copy")
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if me.Hand.Contains(copies[0]) {
		t.Error("the declined copy is still in the hand")
	}
	if !me.Hand.Contains(orig) {
		t.Error("the original left the hand")
	}
}
