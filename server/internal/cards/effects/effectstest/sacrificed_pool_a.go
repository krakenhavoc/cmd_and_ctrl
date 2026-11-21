package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paCast casts a pool spell paying `fodder` and fails the test on a
// refusal.
func PaCast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := CastWithTapParams(t, g, name, typeLine, cost, oracle, params)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// paPermanent is a permanent with a mana cost, for the mana-value
// readers.
func PaPermanent(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}
