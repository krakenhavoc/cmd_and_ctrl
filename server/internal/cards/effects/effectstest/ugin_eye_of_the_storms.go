package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const UginEyeOfTheStormsOracle = "5c58353a-fd60-4528-bf0d-669626cda0b2"

// pushColoredCreature seats a creature with an explicit colour so
// "one or more colors" has something real to read.
func PushUginVictim(g *game.Game, owner uuid.UUID, name string, colors []string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Colors:     colors,
		Owner:      owner,
		Controller: owner,
	})
}

// castUginEyeOfTheStorms puts Ugin in the active seat's hand with his
// PRINTED loyalty stamped on the card — deck import is what supplies
// it in a real game (ADR 0032 §1), and a fixture that skips it seats
// a walker with zero loyalty that the CR 704.5i state-based action
// bins on arrival.
func CastUginEyeOfTheStorms(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID:      id,
		Name:            "Ugin, Eye of the Storms",
		TypeLine:        "Legendary Planeswalker — Ugin",
		OracleID:        UginEyeOfTheStormsOracle,
		StartingLoyalty: 7,
		Owner:           active.ID,
		Controller:      active.ID,
	})
	ToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Ugin: %v", err)
	}
	return id
}
