package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const crossoverCollaborationOracle = "150cefd2-061d-43b2-8470-c17be0144021"

func TestCrossoverCollaborationWithTeamworkAlsoMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	forest, shock, deep := ids[0], ids[1], ids[2]
	before := b16CountNamed(g, "Treasure")
	if _, err := castPaying(t, g, "Crossover Collaboration", "Instant", crossoverCollaborationOracle, nil,
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) || !me.Library.Contains(deep) {
		t.Fatal("the top two cards are exiled")
	}
	for _, id := range []uuid.UUID{forest, shock} {
		perm := exiledPermission(g, id)
		if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
			t.Errorf("%+v: the caster may PLAY it now", perm)
		}
	}
	if got := b16CountNamed(g, "Treasure"); got != before+1 {
		t.Errorf("Treasures = %d, want %d after casting using teamwork", got, before+1)
	}
}

func TestCrossoverCollaborationWithoutTeamworkNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	forest, shock := ids[0], ids[1]
	before := b16CountNamed(g, "Treasure")
	if _, err := castPaying(t, g, "Crossover Collaboration", "Instant", crossoverCollaborationOracle, nil,
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) {
		t.Fatal("the top two cards are exiled")
	}
	if got := b16CountNamed(g, "Treasure"); got != before {
		t.Errorf("Treasures = %d, want %d without teamwork", got, before)
	}
}

func TestCrossoverCollaborationRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weak := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := castPaying(t, g, "Crossover Collaboration", "Instant", crossoverCollaborationOracle, nil,
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
