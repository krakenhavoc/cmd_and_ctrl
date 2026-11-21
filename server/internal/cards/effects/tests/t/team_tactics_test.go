package testst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const teamTacticsOracle = "98b99d7c-d006-471a-8583-0468743c1597"

func TestTeamTacticsWithTeamworkAlsoGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := effectstest.PushVanillaCreature(g, me.ID, "Elf", 1, 1)
	target := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := effectstest.CastPaying(t, g, "Team Tactics", "Instant", teamTacticsOracle, effectstest.TwTarget(target),
		[]int{0}, []uuid.UUID{elf}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	abilities := effectstest.EffectiveAbilities(t, g, target)
	if !effectstest.ContainsString(abilities, "double strike") {
		t.Errorf("double strike not granted")
	}
	if !effectstest.ContainsString(abilities, "trample") {
		t.Errorf("trample not granted when cast using teamwork")
	}
}

func TestTeamTacticsWithoutTeamworkOnlyDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := effectstest.CastPaying(t, g, "Team Tactics", "Instant", teamTacticsOracle, effectstest.TwTarget(target),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	abilities := effectstest.EffectiveAbilities(t, g, target)
	if !effectstest.ContainsString(abilities, "double strike") {
		t.Errorf("double strike not granted")
	}
	if effectstest.ContainsString(abilities, "trample") {
		t.Errorf("trample granted without teamwork")
	}
}

func TestTeamTacticsRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := effectstest.CastPaying(t, g, "Team Tactics", "Instant", teamTacticsOracle, effectstest.TwTarget(target),
		[]int{0}, nil, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
