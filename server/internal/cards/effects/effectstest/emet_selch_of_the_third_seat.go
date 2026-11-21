package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emet_selch_of_the_third_seat_test.go — Emet-Selch of the Third Seat
// (#3785, deck request #3657).

const (
	emetSelchTestOracle = "634475d9-a1d7-4146-9a37-645d3d162af1"
	CauldronFamiliarOra = "a7dc2e62-1c50-4ed7-b71f-2d782a447a5e"
)

func EmetSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	emet := PushDiesCreatureForTest(g, me.ID, "Emet-Selch of the Third Seat", emetSelchTestOracle,
		"Legendary Creature — Elder Wizard", 3, 4)
	ToMainForCost(t, g)
	return g, me, opp, emet
}

// emetOpponentLoses makes opp lose one life outside any resolution.
func EmetOpponentLoses(t *testing.T, g *game.Game, emet uuid.UUID, opp *game.Player) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(emet, opp.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
}

// emetOffer drains opp once, picks target for the trigger and resolves
// it, leaving the cast offer open.
func EmetOffer(t *testing.T, g *game.Game, me, opp *game.Player, emet, target uuid.UUID) {
	t.Helper()
	EmetOpponentLoses(t, g, emet, opp)
	PassPriorityAroundTable(t, g)
	AnswerTriggerTargets(t, g, me.ID, target)
	PassPriorityAroundTable(t, g)
}
