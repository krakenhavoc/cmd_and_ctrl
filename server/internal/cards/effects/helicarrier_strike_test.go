package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const helicarrierStrikeOracle = "ca3cda24-0ecd-4edf-b1b3-54311ad58a51"

// helicarrierAttacker stamps a creature's AttackingTarget directly —
// the target predicate only cares that the field is set, not that a
// full combat declaration produced it.
func helicarrierAttacker(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = defender
				return
			}
		}
	})
}

func TestHelicarrierStrikeWithTeamworkDealsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	helicarrierAttacker(g, victim, me.ID)
	if _, err := castPaying(t, g, "Helicarrier Strike", "Instant", helicarrierStrikeOracle, twTarget(victim),
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := twCard(g, victim).DamageMarked; got != 4 {
		t.Errorf("damage: %d, want 4", got)
	}
}

func TestHelicarrierStrikeWithoutTeamworkDealsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	helicarrierAttacker(g, victim, me.ID)
	if _, err := castPaying(t, g, "Helicarrier Strike", "Instant", helicarrierStrikeOracle, twTarget(victim),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := twCard(g, victim).DamageMarked; got != 2 {
		t.Errorf("damage: %d, want 2", got)
	}
}

func TestHelicarrierStrikeRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	helicarrierAttacker(g, victim, me.ID)
	weak := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := castPaying(t, g, "Helicarrier Strike", "Instant", helicarrierStrikeOracle, twTarget(victim),
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
