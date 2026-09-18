package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_grants_test.go — the catalog half of CR 707.9a / CR 707.9b.
// Phantasmal Image and Sakashima the Impostor are the two cards
// whose except clause grants something, and between them they cover
// all three of it: an added subtype, a granted TRIGGERED ability and
// a granted ACTIVATED one. The case that matters most is the last
// one in the file — a Clone copying a Phantasmal Image gets both,
// which is what makes a grant a copiable value rather than a
// decoration on one permanent.

const (
	oraclePhantasmalImage = "bde94af8-faea-41ff-8eed-ba642eac9968"
)

// hasGrantedSubtype reads the post-layer subtypes, so the assertion
// is about what the rest of the engine sees rather than about a
// string somebody wrote.
func hasGrantedSubtype(c game.Card, s string) bool {
	for _, got := range c.Effective().Subtypes {
		if got == s {
			return true
		}
	}
	return false
}

// castDoomBladeAt casts the catalog's Doom Blade at `victim` from
// the active seat, which is the cheapest way to make something
// "become the target of a spell".
func castDoomBladeAt(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
}

// Asserting the SACRIFICE rather than "it left the battlefield" is the
// point below: Doom Blade would also have destroyed the Image, and
// only one of those is the card's drawback. sacrificedInLog is the
// shared reader for that, in departed_tax_test.go.

// TestPhantasmalImageIsAnIllusionInAdditionToItsOtherTypes —
// CR 707.9b. The copied Bear stays a Bear.
func TestPhantasmalImageIsAnIllusionInAdditionToItsOtherTypes(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Phantasmal Image", "Creature — Illusion", oraclePhantasmalImage, nil)
	resolveWithCopyChoice(t, g, bears)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Grizzly Bears" {
		t.Errorf("name = %q, want the copied creature's", got.Name)
	}
	if got.CurrentPower() != 2 || got.CurrentToughness() != 2 {
		t.Errorf("P/T = %d/%d, want the copied 2/2", got.CurrentPower(), got.CurrentToughness())
	}
	if !hasGrantedSubtype(got, "Illusion") {
		t.Errorf("type line = %q, want the Illusion added", got.TypeLine)
	}
	if !hasGrantedSubtype(got, "Bear") {
		t.Errorf("type line = %q — 'in addition' must not replace the copied subtypes", got.TypeLine)
	}
}

// TestPhantasmalImageIsSacrificedWhenItBecomesATarget — CR 707.9a,
// the granted triggered ability, and the whole cost of the card.
func TestPhantasmalImageIsSacrificedWhenItBecomesATarget(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Phantasmal Image", "Creature — Illusion", oraclePhantasmalImage, nil)
	resolveWithCopyChoice(t, g, bears)

	castDoomBladeAt(t, g, id)
	passPriorityAroundTable(t, g)

	if !sacrificedInLog(g, id) {
		t.Fatal("the Image was not sacrificed — the granted trigger did not fire")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			t.Fatal("the Image is still on the battlefield")
		}
	}
	// The creature it copied is untouched: the grant is on the copy,
	// not on the thing copied.
	if _, still := findBattlefieldByID(g, bears); !still {
		t.Error("the copied Grizzly Bears left the battlefield too")
	}
}

// TestPhantasmalImageWithoutACopyHasNoGrantedTrigger — declining the
// copy means there is no "except" clause, so there is no grant. The
// printed 0/0 Illusion has no sacrifice trigger of its own.
func TestPhantasmalImageWithoutACopyHasNoGrantedTrigger(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Phantasmal Image", "Creature — Illusion", oraclePhantasmalImage, nil)
	resolveWithCopyChoice(t, g, uuid.Nil)

	got := copyBattlefieldCard(t, g, id)
	if len(got.GrantedAbilities) != 0 {
		t.Fatalf("a declined copy carries grants: %v", got.GrantedAbilities)
	}
}

// TestACloneOfAPhantasmalImageInheritsBothHalves is CR 707.9a's
// second sentence and CR 707.9b's: what the except clause granted is
// itself copiable, so the Clone is an Illusion with the trigger.
func TestACloneOfAPhantasmalImageInheritsBothHalves(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	imageID := castCatalogSpell(t, g, "Phantasmal Image", "Creature — Illusion", oraclePhantasmalImage, nil)
	resolveWithCopyChoice(t, g, bears)

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, imageID)

	got := copyBattlefieldCard(t, g, cloneID)
	if !hasGrantedSubtype(got, "Illusion") {
		t.Errorf("the Clone's type line is %q — the added subtype is copiable (CR 707.9b)", got.TypeLine)
	}
	if len(got.GrantedAbilities) == 0 {
		t.Fatal("the Clone did not inherit the granted ability (CR 707.9a)")
	}

	castDoomBladeAt(t, g, cloneID)
	passPriorityAroundTable(t, g)
	if !sacrificedInLog(g, cloneID) {
		t.Fatal("the Clone of the Image was not sacrificed when targeted")
	}
}

// TestSakashimaGrantsItsReturnAbilityToTheCopy — CR 707.9a with an
// ACTIVATED ability, and the caveat this card carried until #665.
func TestSakashimaGrantsItsReturnAbilityToTheCopy(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Sakashima the Impostor",
		"Legendary Creature — Human Rogue", oracleSakashima, nil)
	resolveWithCopyChoice(t, g, bears)

	for i := 0; i < 4; i++ {
		active.ManaPool.AddMana(game.ManaToken{Color: "U"})
	}
	if err := g.ActivateCatalogAbility(active.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the granted activated ability is not there: %v", err)
	}
	passPriorityAroundTable(t, g)

	if _, still := findBattlefieldByID(g, id); !still {
		t.Fatal("Sakashima left immediately — the return is delayed to the next end step (CR 603.7)")
	}

	for i := 0; g.Turn.Step != game.StepEnd; i++ {
		if i >= 12 {
			t.Fatal("never reached the end step")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	if _, still := findBattlefieldByID(g, id); still {
		t.Fatal("Sakashima is still on the battlefield at the end step")
	}
	found := false
	for _, c := range active.Hand.Cards {
		if c.InstanceID == id {
			found = true
			if c.Name != "Sakashima the Impostor" {
				t.Errorf("in hand as %q — CR 400.7, the card reverts to itself", c.Name)
			}
			if len(c.GrantedAbilities) != 0 {
				t.Errorf("the card in hand still carries grants: %v", c.GrantedAbilities)
			}
		}
	}
	if !found {
		t.Fatal("Sakashima is not in its owner's hand")
	}
}

// TestUndoAcrossAPhantasmalImageCopy — the grant has to be part of
// the undo snapshot, and replaying the undone resolution has to land
// the same copy rather than a Phantasmal Image with no grant.
func TestUndoAcrossAPhantasmalImageCopy(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Phantasmal Image", "Creature — Illusion", oraclePhantasmalImage, nil)
	snap := g.Clone()

	resolveWithCopyChoice(t, g, bears)
	if got := copyBattlefieldCard(t, g, id); len(got.GrantedAbilities) == 0 {
		t.Fatal("setup: the copy did not take the grant")
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if _, still := findBattlefieldByID(g, id); still {
		t.Fatal("the undo left the Image on the battlefield")
	}

	resolveWithCopyChoice(t, g, bears)
	replayed := copyBattlefieldCard(t, g, id)
	if len(replayed.GrantedAbilities) == 0 {
		t.Fatalf("after the undo and replay the copy has no grant: %v", replayed.GrantedAbilities)
	}
	if !hasGrantedSubtype(replayed, "Illusion") {
		t.Errorf("after the undo and replay the type line is %q", replayed.TypeLine)
	}
}

// findBattlefieldByID is the by-value battlefield lookup that does
// not fail the test when the card is gone — several assertions here
// are about absence.
func findBattlefieldByID(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
