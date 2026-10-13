package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arcane_bombardment_test.go — #3258, over card_copy_cast_test.go's
// helpers.

// bombardmentGame seats Arcane Bombardment under the active player.
func bombardmentGame(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	enchantment := pushCatalogPermanent(g, me.ID, "Arcane Bombardment", "Enchantment", arcaneBombardmentOracl, false)
	return g, me, opp, enchantment
}

// passUntilCopies passes priority until the trigger has resolved and
// made copies, and stops there with the cast window open.
func passUntilCopies(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 12 && len(cardCopies(g)) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(cardCopies(g)) == 0 {
		t.Fatal("no copies appeared")
	}
}

// dropBoltsFromGraveyard removes the resolved Lightning Bolts the test
// itself cast, so the next random exile has only the seeded cards.
func dropBoltsFromGraveyard(p *game.Player) {
	for _, c := range append([]game.Card(nil), p.Graveyard.Cards...) {
		if c.Name == "Lightning Bolt" {
			_, _ = p.Graveyard.Remove(c.InstanceID)
		}
	}
}

// The first instant or sorcery cast each turn exiles one at random and
// offers a copy of everything exiled with the enchantment.
func TestArcaneBombardmentCopiesTheCardItExiles(t *testing.T) {
	g, me, opp, _ := bombardmentGame(t)
	shock := gyCard(me, "Shock", "Instant", "{R}", shockOracleForCopies)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passUntilCopies(t, g)

	if !inExile(g, shock) {
		t.Fatal("the only instant in the graveyard is exiled at random")
	}
	if n := len(cardCopies(g)); n != 1 {
		t.Fatalf("%d copies, want one for the one exiled card", n)
	}
	cp := copyOf(t, g, "Shock")
	before := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the Shock copy for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3-2 {
		t.Errorf("life = %d, want %d: the Bolt and the copied Shock both resolved", opp.Life, before-5)
	}
	if !inExile(g, shock) || len(cardCopies(g)) != 0 {
		t.Error("the original stays exiled with the enchantment; the copy is gone")
	}
}

// Only the FIRST instant or sorcery each turn triggers; a creature spell
// before it does not count.
func TestArcaneBombardmentOnlyTriggersOnTheFirstInstantOrSorceryEachTurn(t *testing.T) {
	g, me, opp, _ := bombardmentGame(t)
	gyCard(me, "Shock", "Instant", "{R}", shockOracleForCopies)
	gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if len(cardCopies(g)) != 0 {
		t.Fatal("a creature spell does not trigger Arcane Bombardment")
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passUntilCopies(t, g)
	if n := len(cardCopies(g)); n != 1 {
		t.Fatalf("%d copies after the first instant, want 1", n)
	}
	// A second instant the same turn does not trigger again.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if len(cardCopies(g)) != 0 {
		t.Error("copies remain after the window closed")
	}
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Name == "Shock" || c.Name == "Divination" {
			exiled++
		}
	}
	if exiled != 1 {
		t.Errorf("%d graveyard cards exiled, want exactly 1 (the second instant did not trigger)", exiled)
	}
}

// Each trigger exiles ONE more card and copies EVERY card exiled with the
// enchantment, so the copies are made again from the same originals.
func TestArcaneBombardmentCopiesEveryCardExiledWithIt(t *testing.T) {
	g, me, opp, _ := bombardmentGame(t)
	gyCard(me, "Shock", "Instant", "{R}", shockOracleForCopies)
	gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passUntilCopies(t, g)
	if n := len(cardCopies(g)); n != 1 {
		t.Fatalf("first trigger: %d copies, want 1", n)
	}
	firstCopy := cardCopies(g)[0]
	// Decline the first window, then it is "next turn".
	passPriorityAroundTable(t, g)
	if len(cardCopies(g)) != 0 {
		t.Fatal("the declined copy ceased")
	}
	g.WithWriteLock(func() { g.SpellsCastThisTurn = nil })
	dropBoltsFromGraveyard(me)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passUntilCopies(t, g)
	copies := cardCopies(g)
	if len(copies) != 2 {
		t.Fatalf("second trigger: %d copies, want 2 (both exiled cards)", len(copies))
	}
	names := map[string]bool{copies[0].Name: true, copies[1].Name: true}
	if !names["Shock"] || !names["Divination"] {
		t.Errorf("copies = %v, want one of each exiled card", names)
	}
	for _, c := range copies {
		if c.InstanceID == firstCopy.InstanceID {
			t.Error("a copy from the first trigger came back; copies are made anew")
		}
	}
}

// With an empty graveyard the trigger exiles nothing and still copies what
// is already exiled with the enchantment.
func TestArcaneBombardmentWithNothingToExileStillCopiesThePile(t *testing.T) {
	g, me, opp, _ := bombardmentGame(t)
	gyCard(me, "Shock", "Instant", "{R}", shockOracleForCopies)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passUntilCopies(t, g)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.SpellsCastThisTurn = nil })
	dropBoltsFromGraveyard(me)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passUntilCopies(t, g)
	if n := len(cardCopies(g)); n != 1 {
		t.Errorf("%d copies, want 1: the Shock exiled earlier, copied again", n)
	}
}
