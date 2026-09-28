package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const burningCuriosityOracle = "7f420633-901d-47f9-ae0f-0f5b0ea8359c"

func TestBurningCuriosityBlightedExilesThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	forest, shock, bolt, deep := ids[0], ids[1], ids[2], ids[3]
	if _, err := castPaying(t, g, "Burning Curiosity", "Sorcery", burningCuriosityOracle, nil,
		[]int{0}, nil, []uuid.UUID{mine}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) || !g.Exile.Contains(bolt) {
		t.Fatal("the top three cards are exiled when blighted")
	}
	if !me.Library.Contains(deep) {
		t.Error("the fourth card stays in the library")
	}
}

func TestBurningCuriosityDeclinedExilesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	forest, shock, bolt := ids[0], ids[1], ids[2]
	if _, err := castPaying(t, g, "Burning Curiosity", "Sorcery", burningCuriosityOracle, nil,
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest) || !g.Exile.Contains(shock) {
		t.Fatal("the top two cards are exiled")
	}
	if !me.Library.Contains(bolt) {
		t.Error("the third card stays in the library when the cost was declined")
	}
}

func TestBurningCuriosityRefusesBlightWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := castPaying(t, g, "Burning Curiosity", "Sorcery", burningCuriosityOracle, nil,
		[]int{0}, nil, nil); err == nil {
		t.Fatal("expected an error announcing blight with no creature named")
	}
}
