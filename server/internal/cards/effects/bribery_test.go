package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const briberyOracle = "6d194882-ca37-49bb-ac9f-a751c53850a8"

// TestBriberySearchesOpponentsLibraryAndTakesControl proves the
// row's non-owner-searcher / control-changing shape end to end: the
// CASTER is offered the pick from the OPPONENT's library, and the
// creature that arrives enters under the CASTER's control while the
// opponent — not the caster — is the one who shuffles.
func TestBriberySearchesOpponentsLibraryAndTakesControl(t *testing.T) {
	g := newCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]
	toMain(t, g)

	creature := pushLibraryCardForTest(opponent, game.Card{
		Name: "Opponent's Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2,
	})
	land := pushLibraryCardForTest(opponent, game.Card{
		Name: "Opponent's Forest", TypeLine: "Basic Land — Forest",
	})
	casterLibrarySize := len(caster.Library.Cards)

	briberyID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: briberyID, Name: "Bribery", TypeLine: "Sorcery",
		OracleID: briberyOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, briberyID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}},
	}); err != nil {
		t.Fatalf("CastSpell Bribery: %v", err)
	}
	passPriorityAroundTable(t, g)

	// A single creature match with Limit 1 needs no chooser prompt —
	// SearchLibraryThenForEffect's "no decision to make" fast path —
	// so the creature should already be on the battlefield.
	if !g.Battlefield.Contains(creature) {
		t.Fatal("the creature should have entered the battlefield")
	}
	perm, ok := g.LookupCardForEffect(creature)
	if !ok {
		t.Fatal("the entered creature is not on the battlefield")
	}
	if perm.Controller != caster.ID {
		t.Errorf("controller = %v, want the CASTER %v — Bribery puts it under YOUR control", perm.Controller, caster.ID)
	}
	if perm.Owner != opponent.ID {
		t.Errorf("owner = %v, want the searched library's player %v — control changed, ownership did not", perm.Owner, opponent.ID)
	}
	if opponent.Library.Contains(creature) {
		t.Error("the creature should have left the opponent's library")
	}
	if !opponent.Library.Contains(land) {
		t.Error("the untaken land should still be in the opponent's library")
	}
	if got := len(caster.Library.Cards); got != casterLibrarySize {
		t.Errorf("caster's own library size changed (%d -> %d); Bribery never touches it", casterLibrarySize, got)
	}
}

// TestBriberyOnlyOffersCreatureCards is the predicate half: a
// noncreature card in the searched library is never a candidate.
func TestBriberyOnlyOffersCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]
	toMain(t, g)

	creatureA := pushLibraryCardForTest(opponent, game.Card{
		Name: "Bear One", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	creatureB := pushLibraryCardForTest(opponent, game.Card{
		Name: "Bear Two", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	sorcery := pushLibraryCardForTest(opponent, game.Card{
		Name: "A Sorcery", TypeLine: "Sorcery",
	})

	briberyID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: briberyID, Name: "Bribery", TypeLine: "Sorcery",
		OracleID: briberyOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, briberyID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}},
	}); err != nil {
		t.Fatalf("CastSpell Bribery: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Two creature matches for a Limit-1 search is a real choice: the
	// CASTER is offered the pick, over the OPPONENT's library.
	c := searchChoiceFor(g, caster.ID)
	if c == nil {
		t.Fatal("no search prompt with two matching creatures")
	}
	if !hasID(c.SearchCards, creatureA) || !hasID(c.SearchCards, creatureB) {
		t.Errorf("search offers %v, want both creatures", c.SearchCards)
	}
	if hasID(c.SearchCards, sorcery) {
		t.Error("a noncreature card must never be offered")
	}
}

// TestBriberyDoesNotTriggerTheVictimsArchivistOfOghma — #1335. The
// victim's own Archivist of Oghma reads "whenever an OPPONENT
// searches THEIR library" (CR-speak for the searcher's own pile).
// Bribery's caster is an opponent of the victim, but the library
// scanned is the VICTIM's, not the caster's — so this must NOT
// trigger. Before #1335, EventSearchLibrary carried no library-owner
// field at all, so Archivist could not tell "an opponent searched
// their own library" apart from "an opponent searched MY library",
// and this case (which the printed card does not reward) fired.
func TestBriberyDoesNotTriggerTheVictimsArchivistOfOghma(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	toMain(t, g)

	pushCatalogPermanent(g, victim.ID, "Archivist of Oghma",
		"Creature — Halfling Cleric", archivistOfOghmaOracle, false)
	pushLibraryCardForTest(victim, game.Card{
		Name: "Victim's Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	lifeBefore, handBefore := victim.Life, len(victim.Hand.Cards)

	briberyID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: briberyID, Name: "Bribery", TypeLine: "Sorcery",
		OracleID: briberyOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, briberyID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("CastSpell Bribery: %v", err)
	}
	passPriorityAroundTable(t, g)

	if victim.Life != lifeBefore || len(victim.Hand.Cards) != handBefore {
		t.Errorf("victim's Archivist of Oghma triggered off Bribery searching THEIR library (life %d->%d, hand %d->%d); it should only trigger when an opponent searches their OWN library",
			lifeBefore, victim.Life, handBefore, len(victim.Hand.Cards))
	}
}
