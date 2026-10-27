package game

import (
	"testing"

	"github.com/google/uuid"
)

// dead_closure_doors_restore_test.go pins #4361's first slice: four prompts
// that held a frame with nothing in it, and so blocked a restore point for
// no reason, are plain data. A resolution-time colour or creature type
// that nothing follows (it held an empty chooseValueFrame to tell it from
// the stored form), a number or energy payment with no branch (it held a
// closure around nothing), and a library ordering with no continuation (a
// closure around a zero tail). Each test takes a STRICT restore through
// JSON with the prompt open and answers it on the restored table.

// TestATaillessColourChoiceIsARestorePoint: the restored prompt is still
// the resolution-time form, so the answer is announced and stored nowhere.
func TestATaillessColourChoiceIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		g.QueueColorChoiceThenForEffect(ColorPrompt{Chooser: me.ID, Source: source, Question: "choose a color", Purpose: ColorForHarm})
	})
	g = strictRestore(t, g)
	me = g.Seats[0]
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceColor || g.PendingChoices[0].nameThen.Ref != valueChosenNothing {
		t.Fatalf("restored prompts = %+v, want the keyed colour choice", g.PendingChoices)
	}
	before := countEvents(g, EventColorChosen)
	if err := g.ResolveColorChoice(g.PendingChoices[0].ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveColorChoice on the restored table: %v", err)
	}
	if countEvents(g, EventColorChosen) != before+1 {
		t.Error("the restored choice was not announced")
	}
	if got := g.ChosenColorOf(source); got != "" {
		t.Errorf("the resolution-time answer landed on the source as %q, the stored form's job", got)
	}
}

// TestATaillessCreatureTypeChoiceIsARestorePoint is the same for a
// creature type: the source's named tribe stays unset.
func TestATaillessCreatureTypeChoiceIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := pushBear(g, me.ID)
	g.WithWriteLock(func() {
		g.QueueCreatureTypeChoiceThenForEffect(CreatureTypePrompt{Chooser: me.ID, Source: source, Question: "choose a creature type"})
	})
	g = strictRestore(t, g)
	me = g.Seats[0]
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceCreatureType || g.PendingChoices[0].nameThen.Ref != valueChosenNothing {
		t.Fatalf("restored prompts = %+v, want the creature-type choice", g.PendingChoices)
	}
	before := countEvents(g, EventCreatureTypeChosen)
	if err := g.ResolveCreatureTypeChoice(g.PendingChoices[0].ID, me.ID, "elf"); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice on the restored table: %v", err)
	}
	if countEvents(g, EventCreatureTypeChosen) != before+1 {
		t.Error("the restored choice was not announced")
	}
	if i := findCardOnBattlefield(g, source); i < 0 || g.Battlefield.Cards[i].NamedTribe != "" {
		t.Errorf("the resolution-time answer landed on the source, the stored form's job")
	}
}

// TestABranchlessLifePaymentIsARestorePoint: a "pay any amount of life"
// nothing follows holds no frame, and the restored prompt still pays.
func TestABranchlessLifePaymentIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	life := me.Life
	queueChooseNumber(t, g, ChooseNumber{Chooser: me.ID, Source: uuid.New(), Resource: PayResourceLife, Question: "pay any amount of life"})
	g = strictRestore(t, g)
	me = g.Seats[0]
	c := payAmountChoice(g)
	if c == nil {
		t.Fatalf("restored prompts = %+v, want the pay_amount", g.PendingChoices)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err != nil {
		t.Fatalf("ResolvePayAmount on the restored table: %v", err)
	}
	if me.Life != life-3 {
		t.Errorf("life = %d after paying 3 from %d", me.Life, life)
	}
}

// TestABranchlessEnergyPaymentIsARestorePoint is the same for energy.
func TestABranchlessEnergyPaymentIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	setEnergyForTest(t, g, me, 4)
	queuePayAmount(t, g, PayEnergyAmount{Chooser: me.ID, Source: uuid.New(), Question: "pay any amount of energy"})
	g = strictRestore(t, g)
	me = g.Seats[0]
	c := payAmountChoice(g)
	if c == nil {
		t.Fatalf("restored prompts = %+v, want the pay_amount", g.PendingChoices)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 3); err != nil {
		t.Fatalf("ResolvePayAmount on the restored table: %v", err)
	}
	if got := PlayerEnergy(me); got != 1 {
		t.Errorf("energy = %d after paying 3 of 4, want 1", got)
	}
}

// TestAnOrderingNothingFollowsIsARestorePoint: a library ordering with no
// continuation is written, and the restored prompt places the cards.
func TestAnOrderingNothingFollowsIsARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cards := loSeed(me.Hand, me.ID, "First", "Second")
	loQueue(t, g, PutInLibrarySpec{Chooser: me.ID, Cards: cards, From: ZoneHand, Placement: LibraryPlaceTop})
	g = strictRestore(t, g)
	me = g.Seats[0]
	c := loPrompt(t, g)
	if err := g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{cards[1], cards[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary on the restored table: %v", err)
	}
	for _, id := range cards {
		if !me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Errorf("card %s did not move from the hand to the library", id)
		}
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left open: %+v", g.PendingChoices)
	}
}
