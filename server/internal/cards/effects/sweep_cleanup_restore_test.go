package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_cleanup_restore_test.go — #4253's closure sweep, final cleanup: the
// strict mid-prompt restore tests the cards keyed in #4353 went without, and
// one for each prompt this cleanup keyed. Each test opens the prompt, takes a
// STRICT restore through JSON (restoreRoundTrip with strict, which fails on a
// table that is not a restore point) and answers on the restored table.

// sweepCleanupOptionKey returns the open option pick of `chooser`, failing the
// test when it names a key other than `want`.
func sweepCleanupOptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the option pick names key %q, want %q", c.PickThen, want)
	}
	return c
}

// The Gitrog, Ravenous Ride: the "you may sacrifice a creature that saddled
// it" pick and the land put the draw leads to are both restore points, and X
// (the sacrificed creature's power) survives them.
func TestSweepCleanupGitrogRavenousRideIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := pushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", gitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 3)
	lands := pushHandLands(me, 4)
	handBefore := me.Hand.Size()
	gitrogConnects(t, g, me, opp, gitrog, saddler)
	sweep00PickKey(t, g, me.ID, "choose-cards/the-gitrog-ravenous-ride-sacrifice")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/the-gitrog-ravenous-ride-sacrifice")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("sacrifice on the restored table: %v", err)
	}
	if !me.Graveyard.Contains(saddler) {
		t.Fatal("the saddler was not sacrificed on the restored table")
	}
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand %d, want %d (drew X = 3)", got, handBefore+3)
	}
	sweep00PickKey(t, g, me.ID, "choose-cards/put-from-hand-onto-battlefield")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	put := sweep00PickKey(t, g, me.ID, "choose-cards/put-from-hand-onto-battlefield")
	if put.ChooseMax != 3 {
		t.Errorf("the put offers up to %d after the restore, want X = 3", put.ChooseMax)
	}
	if err := g.ResolveChooseCards(put.ID, me.ID, lands[:3]); err != nil {
		t.Fatalf("put lands on the restored table: %v", err)
	}
	for _, id := range lands[:3] {
		if c, ok := battlefieldCardByID(g, id); !ok || !c.Tapped {
			t.Errorf("land %s did not enter tapped on the restored table", id)
		}
	}
}

// Zara, Renegade Recruiter: the pick from the defending player's hand and the
// "what does it attack?" question are restore points, and the creature enters
// attacking what was chosen on the restored table.
func TestSweepCleanupZaraRenegadeRecruiterIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zara := b43Catalog(g, me.ID, "Zara, Renegade Recruiter", "Legendary Creature — Human Pirate", "cd2720c2-522c-4fdc-9cef-9c5ce250fe7b", 4, 3)
	walker := pushWalkerForTest(g, opp.ID, "Test Walker", "", 4)
	beast := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: beast, Name: "Their Beast", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})

	declareAttack(t, g, opp.ID, zara)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/zara-renegade-recruiter-put")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/zara-renegade-recruiter-put")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{beast}); err != nil {
		t.Fatalf("the pick on the restored table: %v", err)
	}
	sweepCleanupOptionKey(t, g, me.ID, "option-pick/zara-renegade-recruiter-attack")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := sweepCleanupOptionKey(t, g, me.ID, "option-pick/zara-renegade-recruiter-attack")
	if len(pick.PickOptions) != 2 {
		t.Fatalf("options %+v, want the player and the planeswalker", pick.PickOptions)
	}
	answerOptionPick(t, g, me.ID, 1)
	got := e2Card(t, g, beast)
	if z := e2Zone(g, beast); z != game.ZoneBattlefield {
		t.Fatalf("the creature is in %q on the restored table, want the battlefield", z)
	}
	if got.Controller != me.ID || !got.Tapped || got.AttackingTarget != walker {
		t.Errorf("controller %s tapped %v attacking %s, want yours, tapped, attacking the planeswalker %s",
			got.Controller, got.Tapped, got.AttackingTarget, walker)
	}
}

// This Is How It Ends: the villainous choice its owner faces and the
// creature pick behind its second option are restore points.
func TestSweepCleanupThisIsHowItEndsIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	wolf := pushVanillaCreature(g, opp.ID, "Wolf", 3, 3)
	libBefore := opp.Library.Size()
	lifeBefore := opp.Life

	villainCastEnds(t, g, bear)
	sweepCleanupOptionKey(t, g, opp.ID, "option-pick/villainous-this-is-how-it-ends")

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	villainAnswer(t, g, sweepCleanupOptionKey(t, g, opp.ID, "option-pick/villainous-this-is-how-it-ends"), 1)
	sweep00PickKey(t, g, opp.ID, "choose-cards/this-is-how-it-ends-shuffle")

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	c := sweep00PickKey(t, g, opp.ID, "choose-cards/this-is-how-it-ends-shuffle")
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{wolf}); err != nil {
		t.Fatalf("the shuffle pick on the restored table: %v", err)
	}
	if findBattlefieldCardForTest(g, wolf) != nil {
		t.Error("the Wolf is still on the battlefield after the restored answer")
	}
	if opp.Library.Size() != libBefore+2 || opp.Life != lifeBefore {
		t.Errorf("library %d -> %d, life %d -> %d; want both creatures shuffled in and no life lost",
			libBefore, opp.Library.Size(), lifeBefore, opp.Life)
	}
}

// Fact or Fiction: the opponent chosen to separate the cards is a keyed
// question now (it carried a closure over the revealed cards), and the split
// it leads to resolves on the restored table.
func TestSweepCleanupFactOrFictionChooseOpponentIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}
		if err := factOrFictionReveal(NewContext(g, item)); err != nil {
			t.Fatalf("Fact or Fiction: %v", err)
		}
	})
	sweepCleanupOptionKey(t, g, me.ID, "option-pick/fact-or-fiction-splitter")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweepCleanupOptionKey(t, g, me.ID, "option-pick/fact-or-fiction-splitter")
	answerOptionPick(t, g, me.ID, 0)
	splitter := c.PickOptions[0].Player
	split := latestRevealPickFor(g, splitter)
	if split == nil {
		t.Fatalf("no split for the chosen opponent %s on the restored table: %+v", splitter, g.PendingChoices)
	}
	if len(split.ChooseCards) != 5 {
		t.Errorf("the split offers %d cards, want the five revealed", len(split.ChooseCards))
	}
}

// Chrome Mox: the imprint pick is the shared keyed exile from a hand, and the
// card it names is exiled on the restored table.
func TestSweepCleanupChromeMoxImprintIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := chromeMoxImprint(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}); err != nil {
			t.Fatalf("imprint: %v", err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/imprint-exile-from-hand")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/imprint-exile-from-hand")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatalf("imprint on the restored table: %v", err)
	}
	if me.Hand.Contains(bolt) || !g.Exile.Contains(bolt) {
		t.Error("the picked card was not exiled on the restored table")
	}
}

// The shared "return the picked graveyard card to its owner's hand" pick
// (Skullwinder's opponent, Ripples of Undeath, Takenuma) is a restore point,
// and the card it names reaches the hand on the restored table.
func TestSweepCleanupReturnPickedToHandIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})
		queueReturnPickedToHand(ctx, me.ID, "test — return a card from your graveyard", []uuid.UUID{dead})
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/return-picked-to-hand")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/return-picked-to-hand")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("the pick on the restored table: %v", err)
	}
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the picked card did not reach its owner's hand on the restored table")
	}
}

// PutFromHandOnTopInAnyOrder (Brainstorm's second sentence): the pick from
// the hand is keyed, and the order prompt it leads to is asked on the
// restored table.
func TestSweepCleanupPutFromHandOnTopIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})
		if err := (PutFromHandOnTopInAnyOrder{N: 2, Label: "test — put two cards back"}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	sweep00PickKey(t, g, me.ID, "choose-cards/put-from-hand-on-top-in-any-order")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/put-from-hand-on-top-in-any-order")
	picked := c.ChooseCards[:2]
	if err := g.ResolveChooseCards(c.ID, me.ID, picked); err != nil {
		t.Fatalf("the pick on the restored table: %v", err)
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatalf("no order prompt after the restored pick: %+v", g.PendingChoices)
	}
	if order.Count != 2 {
		t.Errorf("the order prompt places %d cards, want the two picked", order.Count)
	}
}
