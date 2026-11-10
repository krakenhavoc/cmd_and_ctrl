package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b11_test.go — wave-2 Commander batch 11, the Faeries (#5368).

const (
	tegwyllsScouringOracle = "58ebaaba-7e35-4588-9679-23b22e39789b"
)

func b11Flyer(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := pr7Creature(g, owner, name, 1, "U")
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Keywords = []string{"flying"} })
	return id
}

func b11Tapped(g *game.Game, id uuid.UUID) (tapped bool) {
	g.WithWriteLock(func() { tapped = findBattlefieldCardForTest(g, id).Tapped })
	return
}

func b11Count(g *game.Game, name string) (n int) {
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return
}

// Tegwyll's Scouring at instant speed: tapping three untapped flyers opens
// the window (CR 601.3c), the taps are the cost, and the wrath leaves only
// the three new Faerie Rogues.
func TestTegwyllsScouringFlashByTappingThreeFlyers(t *testing.T) {
	g, me, opp := tapTable(t)
	scour := handCardOf(g, me, "Tegwyll's Scouring", "Sorcery", "{4}{B}{B}", tegwyllsScouringOracle)
	a, b, c := b11Flyer(g, me.ID, "Flyer A"), b11Flyer(g, me.ID, "Flyer B"), b11Flyer(g, me.ID, "Flyer C")
	ground := pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := pr7Creature(g, opp.ID, "Their Bear", 2, "G")
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() { fillPoolColored(me, "B", 6) })

	// Without announcing the cost a sorcery is closed outside a main phase.
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast at instant speed without paying for flash")
	}
	// Two flyers, or a ground creature in the third slot, cannot pay.
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b}}); err == nil {
		t.Fatal("two flyers paid for flash")
	}
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b, ground}}); err == nil {
		t.Fatal("a creature without flying paid for flash")
	}
	if b11Tapped(g, a) || b11Tapped(g, b) {
		t.Fatal("a refused cast tapped a creature")
	}

	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b, c}}); err != nil {
		t.Fatalf("flash cast by tapping three flyers: %v", err)
	}
	if !b11Tapped(g, a) || !b11Tapped(g, b) || !b11Tapped(g, c) {
		t.Fatal("the three flyers were not tapped as a cost")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c, ground, theirs} {
		if g.Battlefield.Contains(id) {
			t.Fatal("a creature survived Tegwyll's Scouring")
		}
	}
	if n := b11Count(g, "Faerie Rogue"); n != 3 {
		t.Fatalf("%d Faerie Rogue tokens, want 3", n)
	}
}

// In a main phase the spell is an ordinary sorcery with no tap cost.
func TestTegwyllsScouringAsASorcery(t *testing.T) {
	g, me, _ := tapTable(t)
	scour := handCardOf(g, me, "Tegwyll's Scouring", "Sorcery", "{4}{B}{B}", tegwyllsScouringOracle)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	g.WithWriteLock(func() { fillPoolColored(me, "B", 6) })
	if err := g.CastSpell(me.ID, scour, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("sorcery cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || b11Count(g, "Faerie Rogue") != 3 {
		t.Fatal("the sweep or the tokens are wrong")
	}
}
