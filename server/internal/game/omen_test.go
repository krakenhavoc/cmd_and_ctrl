package game

import (
	"testing"

	"github.com/google/uuid"
)

// omen_test.go — CR 720 at the engine level (#5175), against a fixture
// with no catalog entry, so what it pins is the LIFECYCLE. The card
// level proofs are the Tarkir: Dragonstorm omen cards' rules tests in
// internal/cards/effects.
//
//	CR 720.3   both halves are offered from hand
//	CR 720.3d  an Omen spell that RESOLVES is shuffled into its owner's
//	           library: not exiled, not in a graveyard, no grant
//	(720.3d's "instead of") countered or fizzled, it is an ordinary
//	           card in an ordinary graveyard
//	CR 720.4   everywhere but the stack the card is its permanent half,
//	           mana value included

// omenFixture is a two-faced omen card as Scryfall lays it out: layout
// `adventure`, a creature front and a sorcery half whose spell type is
// Omen. The Omen half prints {0}, so casting it needs no mana; the
// creature prints {3}, so the two halves' mana values differ.
func omenFixture(owner uuid.UUID) Card {
	c := Card{
		InstanceID: uuid.New(),
		OracleID:   "cccccccc-dddd-eeee-ffff-000000005175",
		Owner:      owner,
		Controller: owner,
		Layout:     LayoutAdventure,
		Faces: []Face{
			{
				Name: "Fixture Regent", TypeLine: "Creature — Dragon",
				ManaCost: "{3}", Colors: []string{"G"},
				Power: 4, Toughness: 4,
			},
			{
				Name: "Fixture Portent", TypeLine: "Sorcery — Omen",
				ManaCost: "{0}", Colors: []string{"G"},
			},
		},
	}
	c.SetFace(0)
	return c
}

// handWithOmen puts an omen fixture in a seat's hand.
func handWithOmen(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	return putInHand(t, g, p, omenFixture(p.ID))
}

func TestOmenHalfIsRecognisedByItsSpellType(t *testing.T) {
	c := omenFixture(uuid.New())
	if castAsOmen(c) {
		t.Error("the creature half reads as an Omen")
	}
	c.SetFace(1)
	if !castAsOmen(c) {
		t.Error("the Omen half is not recognised")
	}
	if castAsAdventure(c) {
		t.Error("an Omen half reads as an Adventure: it would be exiled and castable from exile")
	}
	faces := c.CastableFaces()
	if len(faces) != 2 || faces[0] != 0 || faces[1] != 1 {
		t.Errorf("CastableFaces = %v, want both halves [0 1] (CR 720.3)", faces)
	}

	adv := adventureFixture(uuid.New())
	adv.SetFace(1)
	if castAsOmen(adv) {
		t.Error("an Adventure half reads as an Omen")
	}
}

// CR 720.3d, and the issue's four facts: the library grew by one, the
// card is in it, and it is neither in exile nor in a graveyard. The
// library was shuffled, and the card is its permanent half there.
func TestOmenSpellIsShuffledIntoItsOwnersLibraryOnResolution(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithOmen(t, g, me)
	libBefore := me.Library.Size()

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1, Strict: true}); err != nil {
		t.Fatalf("cast the Omen half: %v", err)
	}
	onStack := findCardForTest(g.Stack, id)
	if onStack == nil || onStack.ActiveFace != 1 || !onStack.IsSorcery() {
		t.Fatalf("the Omen half is not on the stack as face 1: %+v", onStack)
	}
	// CR 720.3b: on the stack it has only the Omen's characteristics.
	if mv := onStack.ManaValue(); mv != 0 {
		t.Errorf("mana value on the stack = %d, want the Omen's 0", mv)
	}
	cut := len(g.Events)
	passPriorityUntilResolvedForTest(t, g, id)

	if !me.Library.Contains(id) {
		t.Fatalf("a resolved Omen is not in its owner's library (exile %v, graveyard %v, hand %v)",
			g.Exile.Contains(id), me.Graveyard.Contains(id), me.Hand.Contains(id))
	}
	if got := me.Library.Size(); got != libBefore+1 {
		t.Errorf("library size = %d, want %d (+1)", got, libBefore+1)
	}
	if g.Exile.Contains(id) {
		t.Error("CR 720.3d: the Omen was exiled like an Adventure")
	}
	if me.Graveyard.Contains(id) {
		t.Error("CR 720.3d: the Omen went to the graveyard")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(id); perm.Granted() {
		t.Errorf("a resolved Omen carries a cast permission %+v", perm)
	}
	if !containsEventKind(g.Events[cut:], EventSearchLibrary) {
		t.Error("the owner's library was not shuffled")
	}
	// CR 720.4: in the library it is the creature, at the creature's
	// mana value.
	in := findCardForTest(me.Library, id)
	if in.ActiveFace != 0 || in.Name != "Fixture Regent" {
		t.Errorf("card in library = face %d %q, want the creature half", in.ActiveFace, in.Name)
	}
	if mv := in.ManaValue(); mv != 3 {
		t.Errorf("mana value in the library = %d, want the creature's 3", mv)
	}
	AssertFaceInvariant(t, g)
}

// The library shuffled is the OWNER's, and the card goes there, even
// when somebody else controlled the spell.
func TestAStolenOmenGoesToItsOwnersLibrary(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	// The active player casts a card the other seat owns.
	id := putInHand(t, g, me, omenFixture(them.ID))
	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Omen half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)
	if !them.Library.Contains(id) {
		t.Fatalf("the Omen is not in its owner's library (caster's library %v)", me.Library.Contains(id))
	}
}

func TestCounteredOmenGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithOmen(t, g, me)
	libBefore := me.Library.Size()

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Omen half: %v", err)
	}
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a countered Omen is not in its owner's graveyard")
	}
	if me.Library.Contains(id) || me.Library.Size() != libBefore {
		t.Error("a countered Omen was shuffled into the library: CR 720.3d is about a spell that resolves")
	}
}

func TestAFizzledOmenGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithOmen(t, g, me)

	gone := uuid.New()
	if err := g.CastSpell(me.ID, id, CastSpellParams{
		Face:    1,
		Targets: []TargetRef{{Kind: TargetCard, ID: gone}},
	}); err != nil {
		t.Fatalf("cast the Omen half at a doomed target: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)
	if !me.Graveyard.Contains(id) {
		t.Fatal("a fizzled Omen is not in its owner's graveyard")
	}
	if me.Library.Contains(id) {
		t.Error("a fizzled Omen was shuffled into the library")
	}
}

// The permanent half is an ordinary creature spell: it resolves onto
// the battlefield, and CR 720.3d has nothing to say about it.
func TestCastingTheOmenCardsPermanentHalfIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	id := handWithOmen(t, g, me)
	if mv := findCardForTest(me.Hand, id).ManaValue(); mv != 3 {
		t.Errorf("mana value in hand = %d, want the creature's 3 (CR 720.4)", mv)
	}
	if err := g.AddManaForEffect(me.ID, uuid.New(), "{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the creature half: %v", err)
	}
	passPriorityUntilResolvedForTest(t, g, id)
	landed := findCardForTest(g.Battlefield, id)
	if landed == nil || landed.ActiveFace != 0 || !landed.IsCreature() {
		t.Fatalf("the creature half did not reach the battlefield: %+v", landed)
	}
	if me.Library.Contains(id) {
		t.Error("the creature half was shuffled away")
	}
}

// CR 616.1: buyback and the Omen shuffle replace the same event, so the
// controller chooses. The prompt is keyed, so the table waiting on it is
// a restore point: answered in the restored game, the Omen shuffle
// still lands the card in its owner's library.
func TestBuybackAndTheOmenShuffleAskTheControllerAcrossARestore(t *testing.T) {
	const oracle = "test-omen-with-buyback"
	stubOptionalCosts(t, CatalogKeyForFace(oracle, adventureSpellFace), []AdditionalCost{
		{Optional: true, Key: BuybackKey, ManaCost: "{0}", Label: "Buyback {0}"},
	})

	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	card := omenFixture(me.ID)
	card.OracleID = oracle
	id := putInHand(t, g, me, card)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Face: 1, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast the Omen half with buyback: %v", err)
	}
	resolveTop(t, g)
	if choice := resolutionChoiceOutstanding(g); choice == nil || len(choice.PickOptions) != 2 {
		t.Fatalf("buyback and the Omen shuffle did not ask the controller: %+v", choice)
	}

	_, restored := roundTrip(t, g)
	choice := resolutionChoiceOutstanding(restored)
	if choice == nil {
		t.Fatal("the restored game lost the prompt")
	}
	// Buyback is offered first; the Omen shuffle second.
	if err := restored.ResolveOptionPick(choice.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	rme := restored.Seats[0]
	if !rme.Library.Contains(id) {
		t.Fatalf("the Omen shuffle did not put the card in the library (hand %v, graveyard %v)",
			rme.Hand.Contains(id), rme.Graveyard.Contains(id))
	}
	if rme.Hand.Contains(id) {
		t.Error("the card is in hand as well — two replacements both applied")
	}
}
