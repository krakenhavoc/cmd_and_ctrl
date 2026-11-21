package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCrossoverCollaborationWithTeamworkAlsoMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	forest, shock, deep := ids[0], ids[1], ids[2]
	before := effectstest.B16CountNamed(g, "Treasure")
	if _, err := effectstest.CastPaying(t, g, "Crossover Collaboration", "Instant", effectstest.CrossoverCollaborationOracle, nil,
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) || !me.Library.Contains(deep) {
		t.Fatal("the top two cards are exiled")
	}
	for _, id := range []uuid.UUID{forest, shock} {
		perm := effectstest.ExiledPermission(g, id)
		if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
			t.Errorf("%+v: the caster may PLAY it now", perm)
		}
	}
	if got := effectstest.B16CountNamed(g, "Treasure"); got != before+1 {
		t.Errorf("Treasures = %d, want %d after casting using teamwork", got, before+1)
	}
}

func TestCrossoverCollaborationWithoutTeamworkNoTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	forest, shock := ids[0], ids[1]
	before := effectstest.B16CountNamed(g, "Treasure")
	if _, err := effectstest.CastPaying(t, g, "Crossover Collaboration", "Instant", effectstest.CrossoverCollaborationOracle, nil,
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) {
		t.Fatal("the top two cards are exiled")
	}
	if got := effectstest.B16CountNamed(g, "Treasure"); got != before {
		t.Errorf("Treasures = %d, want %d without teamwork", got, before)
	}
}

func TestCrossoverCollaborationRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weak := effectstest.PushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := effectstest.CastPaying(t, g, "Crossover Collaboration", "Instant", effectstest.CrossoverCollaborationOracle, nil,
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
