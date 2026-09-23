package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// life_cost_lock_test.go — the enumerator half of #1200 (CR 119.8,
// CR 614.17b), in the #499/#618 agreement style: every activation the
// enumerator OFFERS is one the engine accepts, and every activation
// the engine REFUSES is one the enumerator does not offer.
//
// That agreement is the whole reason both read game.CanPayLifeLocked.
// A bot offered a life cost its locked seat cannot pay picks it, is
// refused with ErrInvalidParam, and picks it again.

const oracleLifeCostAltar = "test-legal-life-cost-altar"

// altarWithALifeCost seats a permanent whose one ability costs `n`
// life and nothing else — the shape of every "pay N life:" activation
// in the catalog, with the mana and tap components left out so the
// life gate is the only thing that can remove the move.
func altarWithALifeCost(g *game.Game, p *game.Player, n int) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name:     "Blood Altar",
		TypeLine: "Artifact",
		OracleID: oracleLifeCostAltar,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Pay life: draw nothing",
			Cost:  game.AbilityCost{Life: n},
			Effect: func(_ *game.Game, _ *game.StackItem) error {
				return nil
			},
		}},
	})
}

// TestALifeCostIsNotOfferedToALockedSeat — CR 119.8 makes the cost
// unpayable, so CR 614.17b makes it unchoosable, so the enumerator
// must not list it.
func TestALifeCostIsNotOfferedToALockedSeat(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	altar := altarWithALifeCost(g, active, 2)
	advanceTo(t, g, game.StepPrecombatMain)

	// Unlocked the ability is a move — otherwise its absence below
	// would prove nothing.
	if moves := legal.EnumerateFor(g, active.ID); len(activationsOf(moves, altar)) == 0 {
		t.Fatal("a payable life cost should be a move with nothing locking the seat")
	}

	g.WithWriteLock(func() {
		g.GrantLifeTotalLockForEffect(active.ID, "Test — your life total can't change",
			uuid.Nil, g.UntilYourNextTurnDuration(active.ID))
	})

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), altar); len(acts) != 0 {
		t.Errorf("a life cost is still offered to a seat whose life total can't change: %v", labels(acts))
	}
	// And the engine agrees, which is the half that makes the absence
	// correct rather than merely conservative.
	if err := g.ActivateCatalogAbility(active.ID, altar, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the engine accepted a life cost the enumerator refused to offer")
	}
}

// TestAZeroLifeCostIsStillOfferedToALockedSeat — Platinum Emperion's
// own parenthetical: "You can't pay any amount of life except 0". A
// predicate that refused every ability with a Life field rather than
// a positive payment would fail here.
func TestAZeroLifeCostIsStillOfferedToALockedSeat(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	altar := altarWithALifeCost(g, active, 0)
	advanceTo(t, g, game.StepPrecombatMain)

	g.WithWriteLock(func() {
		g.GrantLifeTotalLockForEffect(active.ID, "Test — your life total can't change",
			uuid.Nil, g.UntilYourNextTurnDuration(active.ID))
	})
	if moves := legal.EnumerateFor(g, active.ID); len(activationsOf(moves, altar)) == 0 {
		t.Error("an ability that costs no life is not a payment, and must still be offered")
	}
}

// TestALockOnOneSeatDoesNotRemoveAnotherSeatsLifeCost is the control
// the reader could otherwise get wrong in the cheapest possible way:
// reading the board rather than the seat.
func TestALockOnOneSeatDoesNotRemoveAnotherSeatsLifeCost(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	altar := altarWithALifeCost(g, active, 2)
	advanceTo(t, g, game.StepPrecombatMain)

	g.WithWriteLock(func() {
		g.GrantLifeTotalLockForEffect(other.ID, "Test — their life total can't change",
			uuid.Nil, g.UntilYourNextTurnDuration(other.ID))
	})
	if moves := legal.EnumerateFor(g, active.ID); len(activationsOf(moves, altar)) == 0 {
		t.Error("an opponent's lock removed this seat's life cost from the enumeration")
	}
}
