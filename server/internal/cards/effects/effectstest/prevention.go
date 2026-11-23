package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention_test.go — CR 615.8, the charged shield. fog_test.go
// already pins the uncharged Fog shape; what is new in S30 is the
// arithmetic, and the arithmetic has exactly three interesting
// cases: the shield covers the damage, the shield is smaller than
// the damage, and the shield has already been spent.
//
// The middle one is the case the generous first draft gets wrong. A
// 4-point shield facing 6 damage must let 2 through; an
// implementation that cancels the event whenever the shield covers
// any of it prevents all 6 and passes every test that only checks
// "did the creature survive".

const (
	HolyDayOracle     = "98423a34-f044-4811-b288-56981d604b6e"
	TangleOracle      = "f627e125-15af-4e53-b34e-82b60e4ec87b"
	MendingHandsOracl = "a612f30d-cd55-438b-a7de-8c80509183aa"
)

func PushBear(g *game.Game, owner uuid.UUID, name string, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: toughness, Owner: owner, Controller: owner,
	})
	return id
}

func LifeTotalOf(g *game.Game, playerID uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == playerID {
			return p.Life
		}
	}
	return -1
}

func DamageMarkedOn(g *game.Game, id uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}
