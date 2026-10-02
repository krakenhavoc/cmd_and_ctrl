package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// until_return_test.go — #1729, CR 610.3: an exile "until" an event is
// returned by a one-shot effect immediately after the event, never by
// a trigger on the stack, and it does not leave the game with the
// player whose ability exiled the card.

// untilReturnOnStack reports whether anything is waiting to resolve or
// to be put on the stack — the window CR 610.3 does not open.
func untilReturnOnStack(g *game.Game) bool {
	busy := false
	g.ReadSnapshot(func() { busy = !stackFullyEmpty(g) })
	return busy
}

// untilRecords counts the "until" records still owed.
func untilRecords(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, dt := range g.DelayedTriggers {
			if dt != nil && dt.Until {
				n++
			}
		}
	})
	return n
}

// --- Palace Jailer ---------------------------------------------------

// TestPalaceJailerControllerLeavingReturnsTheCreature is the issue's
// first item: the Jailer's controller concedes wearing the crown, CR
// 725.4 hands it to an opponent of theirs as they leave, and that is
// the printed "until". The return is a one-shot effect (CR 610.3), so
// it does not leave the game with them (CR 800.4a/d name triggered
// abilities).
func TestPalaceJailerControllerLeavingReturnsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailerExile(t, g, victim)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.Monarch == uuid.Nil || g.Monarch == me.ID {
		t.Fatalf("CR 725.4: monarch = %v, want a player still in the game", g.Monarch)
	}
	back, ok := e2BattlefieldNamed(g, "Their Bear")
	if !ok {
		t.Fatal("the Jailer's controller left, an opponent of theirs took the crown, and the creature stayed in exile")
	}
	if back.Controller != owner.ID {
		t.Errorf("returned under %v, want its owner %v (CR 610.3c)", back.Controller, owner.ID)
	}
	if untilRecords(g) != 0 {
		t.Errorf("%d until records left after the return", untilRecords(g))
	}
}

// TestPalaceJailerControllerLeavingWithoutTheCrownKeepsWatching: a
// controller who leaves while nobody is the monarch hands nothing on
// (CR 725.4 has no monarch to replace), so the creature stays exiled —
// and the game keeps watching. The next opponent of theirs to become
// the monarch brings it back: the record is not a trigger, so it did
// not leave with them, and "an opponent" is read by last-known
// information (CR 800.4i).
func TestPalaceJailerControllerLeavingWithoutTheCrownKeepsWatching(t *testing.T) {
	g := newCatalogGame(t)
	me, owner, thief := g.Seats[0], g.Seats[1], g.Seats[2]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailerExile(t, g, victim)
	monCrown(t, g, uuid.Nil)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature is in %q; nobody became the monarch", z)
	}
	monCrown(t, g, thief.ID)
	if _, ok := e2BattlefieldNamed(g, "Their Bear"); !ok {
		t.Error("an opponent of the departed Jailer controller became the monarch and the creature stayed exiled")
	}
}

// TestPalaceJailerReturnUsesNoStack is the issue's third item: CR 610.3
// creates the return "immediately after the specified event", so there
// is no window in which the creature is still in exile.
func TestPalaceJailerReturnUsesNoStack(t *testing.T) {
	g := newCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	jailerExile(t, g, victim)
	monCrown(t, g, thief.ID)
	if _, ok := e2BattlefieldNamed(g, "Their Bear"); !ok {
		t.Fatal("the crown moved to an opponent and the creature is not back yet")
	}
	if untilReturnOnStack(g) {
		t.Error("the return put something on the stack")
	}
}

// TestPalaceJailerOpponentCrownedInResponseExilesNothing is CR 610.3b:
// an opponent who BECOMES the monarch after the exile ability triggered
// and before it resolves has already ended the "until", so the creature
// does not move at all.
func TestPalaceJailerOpponentCrownedInResponseExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, owner, thief := g.Seats[0], g.Seats[1], g.Seats[2]
	victim := pushVanillaCreature(g, owner.ID, "Their Bear", 2, 2)
	castAndResolveCreature(t, g, "Palace Jailer", "Creature — Human Soldier", palaceJailerOracle)
	pickCard(t, g, me.ID, victim)
	// Throne of the High City in response, as far as the rules care.
	monCrown(t, g, thief.ID)
	monSettle(t, g)
	if z := e2Zone(g, victim); z != game.ZoneBattlefield {
		t.Errorf("the creature is in %q; an opponent became the monarch before the exile resolved (CR 610.3b)", z)
	}
	if untilRecords(g) != 0 {
		t.Errorf("%d until records for an exile that never happened", untilRecords(g))
	}
}

// --- Hostage Taker ---------------------------------------------------

func takerHoldsVictim(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	taker := castAndResolveCreature(t, g, "Hostage Taker", "Creature — Human Pirate", hostageTakerOracle)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature is in %q, want exile", z)
	}
	return taker
}

// TestHostageTakerReturnsImmediatelyWhenItDies: "Nothing happens
// between the two events, including state-based actions" (ruling
// 2017-09-29). The Taker is destroyed, and the card is back before
// anything reaches the stack.
func TestHostageTakerReturnsImmediatelyWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	taker := takerHoldsVictim(t, g, victim)
	// Destroyed by an effect, then the CR 704.3 boundary every
	// resolution ends at — and nothing else.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(taker) })
	g.RunStateChecksForTest()
	if z := e2Zone(g, taker); z != game.ZoneGraveyard {
		t.Fatalf("the Taker is in %q, want the graveyard", z)
	}
	back, ok := e2BattlefieldNamed(g, "Their Bear")
	if !ok {
		t.Fatal("the Taker died and the card did not return before anyone could act")
	}
	if back.Controller != opp.ID {
		t.Errorf("returned under %v, want its owner %v", back.Controller, opp.ID)
	}
	if untilReturnOnStack(g) {
		t.Error("the return went through the stack")
	}
}

// TestHostageTakerOwnerLeavingReturnsTheCard is the ruling of
// 2017-09-29: "if Hostage Taker's owner leaves the game while the card
// is still exiled and another player owns that card, the exiled card
// will return to the battlefield under its owner's control."
func TestHostageTakerOwnerLeavingReturnsTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	takerHoldsVictim(t, g, victim)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	back, ok := e2BattlefieldNamed(g, "Their Bear")
	if !ok {
		t.Fatal("the Taker left the game with its owner and the card stayed in exile")
	}
	if back.Controller != opp.ID {
		t.Errorf("returned under %v, want its owner %v", back.Controller, opp.ID)
	}
}

// TestHostageTakerReturnSurvivesARestore: the record is data, so a
// table with a card held is a restore point, and the restored game
// still returns it.
func TestHostageTakerReturnSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := ctrlPushCreature(g, opp.ID, "Their Bear")
	taker := takerHoldsVictim(t, g, victim)
	restored := restoreRoundTrip(t, g, true)
	if untilRecords(restored) != 1 {
		t.Fatalf("restored %d until records, want 1", untilRecords(restored))
	}
	restored.WithWriteLock(func() { _ = restored.DestroyPermanentForEffect(taker) })
	restored.RunStateChecksForTest()
	if _, ok := e2BattlefieldNamed(restored, "Their Bear"); !ok {
		t.Error("the restored game did not return the card when the Taker died")
	}
}

// --- Ossification ----------------------------------------------------

func ossificationExile(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	plains := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	aura := castCatalogSpell(t, g, "Ossification", "Enchantment — Aura", b41OssificationOracle, cardRefs(plains))
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature is in %q, want exile", z)
	}
	return aura
}

// TestOssificationThatLeftFirstExilesNothing is CR 610.3b and the
// ruling of 2023-02-04: "If Ossification leaves the battlefield before
// its triggered ability resolves, the target permanent won't be
// exiled." Before #1729 the card was exiled for good: its leave trigger
// had already gone by.
func TestOssificationThatLeftFirstExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	aura := castCatalogSpell(t, g, "Ossification", "Enchantment — Aura", b41OssificationOracle, cardRefs(plains))
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	if triggerOnStack(g, aura) == nil {
		t.Fatal("the exile trigger is not on the stack")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, victim); z != game.ZoneBattlefield {
		t.Errorf("the creature is in %q; an Ossification that had already left exiled it", z)
	}
}

// TestOssificationOwnerLeavingReturnsTheCard: the Aura leaves the
// battlefield with its owner (CR 800.4a), and that is the event its
// "until" waits for, though no zone-change event announces it.
func TestOssificationOwnerLeavingReturnsTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	ossificationExile(t, g, victim)
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if namedOnBattlefield(g, "Their Threat", opp.ID) != 1 {
		t.Error("the Aura left the game with its owner and the creature stayed in exile")
	}
}

// TestOssificationLegacyExileStillReturns: a card exiled by a binary
// from before #1729 has no until record. The legacy leave trigger the
// card kept brings it back.
func TestOssificationLegacyExileStillReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ossification", TypeLine: "Enchantment — Aura",
		OracleID: b41OssificationOracle, Owner: me.ID, Controller: me.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: plains},
	})
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	// The old shape: the exile resolved under the entry trigger's
	// label, and nothing recorded a return.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Source: aura, Label: b41OssificationExileLabel})
		if err := g.ExileCardForEffect(victim); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	if untilRecords(g) != 0 {
		t.Fatal("the legacy exile has a record")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	passPriorityAroundTable(t, g)
	if namedOnBattlefield(g, "Their Threat", opp.ID) != 1 {
		t.Error("a card exiled before #1729 did not come back when the Aura left")
	}
}

// TestOssificationNewExileDoesNotFireTheLegacyTrigger: with a record in
// place, the leave trigger never reaches the stack.
func TestOssificationNewExileDoesNotFireTheLegacyTrigger(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	aura := ossificationExile(t, g, victim)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	if pending := len(g.PendingTriggers); pending != 0 {
		t.Errorf("%d triggers queued by the Aura leaving; the return is not a trigger", pending)
	}
	g.RunStateChecksForTest()
	if namedOnBattlefield(g, "Their Threat", opp.ID) != 1 {
		t.Error("the creature did not come back when the Aura left")
	}
}
