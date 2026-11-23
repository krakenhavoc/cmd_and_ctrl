package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestNarcomoebaReturnsItselfWhenMilled: the card is in the graveyard
// when the ability triggers — being put there IS the trigger — so
// this fires only because the harvest walks the graveyard for a card
// that declared it (#925).
func TestNarcomoebaReturnsItselfWhenMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PutOnTopOfLibrary(g, me, "Narcomoeba", "Creature — Illusion", effectstest.NarcomoebaOracle)

	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	if !effectstest.InZone(me.Graveyard, id) {
		t.Fatal("the mill did not put Narcomoeba into its owner's graveyard")
	}

	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.InZone(me.Graveyard, id) {
		t.Error("Narcomoeba is still in the graveyard after its trigger resolved")
	}
	if !effectstest.OnBattlefieldByName(g, "Narcomoeba") {
		t.Error("Narcomoeba did not reach the battlefield")
	}
}

// TestNarcomoebaDeclinedStaysInTheGraveyard: the "you may" is a CR
// 603.5 prompt, so "no" drops the trigger and leaves the card where
// it is. It also proves the graveyard trigger takes the ORDINARY
// dispatch rather than a second path of its own.
func TestNarcomoebaDeclinedStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PutOnTopOfLibrary(g, me, "Narcomoeba", "Creature — Illusion", effectstest.NarcomoebaOracle)

	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InZone(me.Graveyard, id) {
		t.Error("a declined Narcomoeba left the graveyard anyway")
	}
}

// TestNarcomoebaOnTheBattlefieldDoesNotWatchAMill is the negative
// half, and the assertion that fails with the fix backed out: a
// Narcomoeba already on the battlefield has no graveyard ability,
// and before the Zones dimension the harvester's battlefield walk
// was the ONLY walk — so a card declaring this trigger would have
// fired from play and tried to return itself from a graveyard it is
// not in.
func TestNarcomoebaOnTheBattlefieldDoesNotWatchAMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(),
		Name:       "Narcomoeba",
		TypeLine:   "Creature — Illusion",
		OracleID:   effectstest.NarcomoebaOracle,
		Power:      1,
		Toughness:  1,
		Owner:      me.ID,
		Controller: me.ID,
	})
	effectstest.PutOnTopOfLibrary(g, me, "Filler", "Creature — Test", "")

	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})

	if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
		t.Fatalf("a battlefield Narcomoeba queued a graveyard trigger (%q)", c.Reason)
	}
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("a battlefield Narcomoeba queued %d triggers, want 0", len(g.PendingTriggers))
	}
}

// TestBloodghastReturnsOnLandfallFromTheGraveyard: landfall watched
// from a graveyard, with "you control" and "your graveyard" both
// meaning the card's OWNER (CR 108.4).
func TestBloodghastReturnsOnLandfallFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       "Bloodghast",
		TypeLine:   "Creature — Vampire Spirit",
		OracleID:   effectstest.BloodghastOracle,
		Power:      2,
		Toughness:  1,
		Owner:      me.ID,
	})

	effectstest.PlayLandFromHand(t, g, "Swamp", "")

	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.InZone(me.Graveyard, id) {
		t.Error("Bloodghast is still in the graveyard after landfall")
	}
	if !effectstest.OnBattlefieldByName(g, "Bloodghast") {
		t.Error("Bloodghast did not return to the battlefield")
	}
}

// TestBloodghastIgnoresAnOpponentsLand: "a land YOU control", read
// off the graveyard card's OWNER. An opponent's land drop is not
// Bloodghast's landfall.
func TestBloodghastIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	me.Graveyard.PushTop(game.Card{
		InstanceID: uuid.New(),
		Name:       "Bloodghast",
		TypeLine:   "Creature — Vampire Spirit",
		OracleID:   effectstest.BloodghastOracle,
		Owner:      me.ID,
	})

	// The opponent's land enters directly — no land-drop bookkeeping
	// needed, the trigger watches the ENTRY.
	g.WithWriteLock(func() {
		landID := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: landID,
			Name:       "Island",
			TypeLine:   "Basic Land — Island",
			Owner:      opponent.ID,
			Controller: opponent.ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: landID, Actor: opponent.ID})
	})

	if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
		t.Fatalf("an opponent's land woke Bloodghast (%q)", c.Reason)
	}
}

// Two landfall triggers on the stack (two lands entering, or a fetch
// and its land): the first returns Bloodghast, and the second finds it
// gone from the graveyard. CR 400.7 makes that a new object, so the
// second does nothing — it used to fail with "card instance not found
// in zone", which the catalog soak caught.
func TestBloodghastSecondLandfallAfterItReturnedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedPermanentWithOracle(g, me.ID, "Bloodghast", "Creature — Vampire Spirit", effectstest.BloodghastOracle)
	var err error
	g.WithWriteLock(func() {
		err = effects.ReturnThisCardFromYourGraveyard(g, &game.StackItem{
			Kind: game.StackItemTriggered, SourceCardID: id, Controller: me.ID,
		})
	})
	if err != nil {
		t.Fatalf("a landfall resolving after Bloodghast left the graveyard: %v", err)
	}
	if !effectstest.OnBattlefieldByName(g, "Bloodghast") {
		t.Error("Bloodghast moved")
	}
}
