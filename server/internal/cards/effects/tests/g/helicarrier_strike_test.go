package testsg

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const helicarrierStrikeOracle = "ca3cda24-0ecd-4edf-b1b3-54311ad58a51"

func TestHelicarrierStrikeWithTeamworkDealsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := effectstest.PushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	effectstest.HelicarrierAttacker(g, victim, me.ID)
	if _, err := effectstest.CastPaying(t, g, "Helicarrier Strike", "Instant", helicarrierStrikeOracle, effectstest.TwTarget(victim),
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TwCard(g, victim).DamageMarked; got != 4 {
		t.Errorf("damage: %d, want 4", got)
	}
}

func TestHelicarrierStrikeWithoutTeamworkDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	effectstest.HelicarrierAttacker(g, victim, me.ID)
	if _, err := effectstest.CastPaying(t, g, "Helicarrier Strike", "Instant", helicarrierStrikeOracle, effectstest.TwTarget(victim),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TwCard(g, victim).DamageMarked; got != 2 {
		t.Errorf("damage: %d, want 2", got)
	}
}

func TestHelicarrierStrikeRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	effectstest.HelicarrierAttacker(g, victim, me.ID)
	weak := effectstest.PushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := effectstest.CastPaying(t, g, "Helicarrier Strike", "Instant", helicarrierStrikeOracle, effectstest.TwTarget(victim),
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
