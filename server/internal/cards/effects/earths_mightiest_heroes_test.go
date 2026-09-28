package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const earthsMightiestHeroesOracle = "2b87bcac-ec64-4012-9e75-f459d3640eb3"

func TestEarthsMightiestHeroesWithTeamworkPutsBothCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Elf A", 5, 5)
	b := pushVanillaCreature(g, me.ID, "Elf B", 5, 5)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear One", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Bear Two", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	bearOne, bearTwo, shock := ids[0], ids[1], ids[2]
	if _, err := castPaying(t, g, "Earth's Mightiest Heroes", "Sorcery", earthsMightiestHeroesOracle, nil,
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Earth's Mightiest Heroes asked nothing")
	}
	if pick.ChooseMax < 2 {
		t.Errorf("ChooseMax = %d, want at least 2 (any number) when cast using teamwork", pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, bearOne, bearTwo)
	if !g.Battlefield.Contains(bearOne) || !g.Battlefield.Contains(bearTwo) {
		t.Error("both creature cards should have entered the battlefield")
	}
	if !me.Graveyard.Contains(shock) {
		t.Error("the non-creature card should be in the graveyard")
	}
}

func TestEarthsMightiestHeroesWithoutTeamworkPutsOnlyOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear One", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Bear Two", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	bearOne, bearTwo, shock := ids[0], ids[1], ids[2]
	if _, err := castPaying(t, g, "Earth's Mightiest Heroes", "Sorcery", earthsMightiestHeroesOracle, nil,
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Earth's Mightiest Heroes asked nothing")
	}
	if pick.ChooseMax != 1 {
		t.Errorf("ChooseMax = %d, want 1 without teamwork", pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, bearOne)
	if !g.Battlefield.Contains(bearOne) {
		t.Error("the chosen creature card should have entered the battlefield")
	}
	if !me.Graveyard.Contains(bearTwo) || !me.Graveyard.Contains(shock) {
		t.Error("the rest should be in the graveyard")
	}
}

func TestEarthsMightiestHeroesRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weak := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := castPaying(t, g, "Earth's Mightiest Heroes", "Sorcery", earthsMightiestHeroesOracle, nil,
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
