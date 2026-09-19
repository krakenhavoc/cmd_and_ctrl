package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const improvisationCapstoneOracle = "fd4f0315-567a-4cc5-bc7e-88a9d2cee910"

// improvisationCapstoneCard is a library card with a printed mana
// cost, which is what the running total reads.
func improvisationCapstoneCard(name, typeLine, manaCost string) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, ManaCost: manaCost}
}

// "Until you exile cards with total mana value 4 or greater": the
// card that carries the total over the line is exiled too, and the
// exile stops there.
func TestImprovisationCapstoneExilesUntilTheTotalReachesFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// seedSearchLibrary pushes each under the previous, so the FIRST
	// listed is on top.
	ids := seedSearchLibrary(me,
		improvisationCapstoneCard("One", "Instant", "{R}"),
		improvisationCapstoneCard("Two", "Sorcery", "{1}{R}"),
		improvisationCapstoneCard("Three", "Creature — Bear", "{2}{R}"),
		improvisationCapstoneCard("Deep", "Instant", "{R}"),
	)
	castCatalogSpell(t, g, "Improvisation Capstone", "Sorcery — Lesson", improvisationCapstoneOracle, nil)
	passPriorityAroundTable(t, g)

	// 1 + 2 = 3, still short; the third card takes it to 6 and the
	// exile stops. The fourth stays in the library.
	for i, want := range []bool{true, true, true, false} {
		if g.Exile.Contains(ids[i]) != want {
			t.Errorf("card %d exiled=%v, want %v", i, !want, want)
		}
	}
	if me.Library.Size() != 1 {
		t.Errorf("library %d, want 1", me.Library.Size())
	}
}

// The permission is exactly what is printed: the caster may CAST the
// exiled cards, at no mana cost, with no stated expiry.
func TestImprovisationCapstoneGrantsThePrintedPermission(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		improvisationCapstoneCard("Bolt", "Instant", "{R}"),
		improvisationCapstoneCard("Wurm", "Creature — Wurm", "{5}{G}"),
	)
	castCatalogSpell(t, g, "Improvisation Capstone", "Sorcery — Lesson", improvisationCapstoneOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range ids {
		if !g.Exile.Contains(id) {
			t.Fatalf("%s reached exile", id)
		}
		perm := exiledPermission(g, id)
		if perm.Player != me.ID {
			t.Errorf("the caster holds the grant, got %v", perm.Player)
		}
		if perm.Cost != "{0}" {
			t.Errorf("no mana is charged, got cost %q", perm.Cost)
		}
		if !perm.CastOnly {
			t.Error(`"cast any number of spells" does not let a land be played`)
		}
		if perm.Duration.Kind != game.Indefinite {
			t.Errorf("the card states no window: duration %v", perm.Duration.Kind)
		}
	}

	// Any NUMBER of them: casting one leaves the other castable.
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("casting the first exiled card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !permissionLive(g, exiledPermission(g, ids[1]), me.ID) {
		t.Error("the second card is still castable after the first was cast")
	}
}

// An empty library is not an error: nothing is exiled and nothing is
// granted.
func TestImprovisationCapstoneSurvivesAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	castCatalogSpell(t, g, "Improvisation Capstone", "Sorcery — Lesson", improvisationCapstoneOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life <= 0 {
		t.Error("nobody loses for an empty library here — only a draw does")
	}
}

// The grants are per OBJECT, at the epoch the card was exiled at
// (ADR 0066, CR 400.7): one row per exiled card, and a card that
// leaves exile and comes back is a new object with nothing granted.
func TestImprovisationCapstoneGrantsAreOneObjectEach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		improvisationCapstoneCard("Bolt", "Instant", "{R}"),
		improvisationCapstoneCard("Wurm", "Creature — Wurm", "{5}{G}"),
	)
	castCatalogSpell(t, g, "Improvisation Capstone", "Sorcery — Lesson", improvisationCapstoneOracle, nil)
	passPriorityAroundTable(t, g)

	if len(me.CastPermissions) != len(ids) {
		t.Fatalf("%d grants for %d exiled cards — one each", len(me.CastPermissions), len(ids))
	}
	for _, perm := range me.CastPermissions {
		if len(perm.Cards) != 1 {
			t.Errorf("a grant names %d objects, want 1", len(perm.Cards))
		}
	}

	// Out of exile and back again: a new object, nothing granted.
	g.WithWriteLock(func() {
		_ = g.BounceToHandForEffect(ids[1])
		_, _ = game.MoveCard(me.Hand, g.Exile, ids[1])
	})
	if permissionLive(g, exiledPermission(g, ids[1]), me.ID) {
		t.Error("a card that left exile and came back is a new object with nothing granted")
	}
}
