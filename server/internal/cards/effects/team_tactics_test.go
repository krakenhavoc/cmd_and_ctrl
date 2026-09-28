package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const teamTacticsOracle = "98b99d7c-d006-471a-8583-0468743c1597"

func TestTeamTacticsWithTeamworkAlsoGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	target := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castPaying(t, g, "Team Tactics", "Instant", teamTacticsOracle, twTarget(target),
		[]int{0}, []uuid.UUID{elf}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	abilities := effectiveAbilities(t, g, target)
	if !containsString(abilities, "double strike") {
		t.Errorf("double strike not granted")
	}
	if !containsString(abilities, "trample") {
		t.Errorf("trample not granted when cast using teamwork")
	}
}

func TestTeamTacticsWithoutTeamworkOnlyDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castPaying(t, g, "Team Tactics", "Instant", teamTacticsOracle, twTarget(target),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	abilities := effectiveAbilities(t, g, target)
	if !containsString(abilities, "double strike") {
		t.Errorf("double strike not granted")
	}
	if containsString(abilities, "trample") {
		t.Errorf("trample granted without teamwork")
	}
}

func TestTeamTacticsRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castPaying(t, g, "Team Tactics", "Instant", teamTacticsOracle, twTarget(target),
		[]int{0}, nil, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
