package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ebPush puts a catalog permanent onto me's battlefield, not summoning
// sick.
func EbPush(g *game.Game, me *game.Player, name, oracle, typeLine string, power, toughness int) uuid.UUID {
	return ApaPush(g, me.ID, me.ID, game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: power, Toughness: toughness})
}

func EbCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func EbLive(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	c := ApaLive(g, id)
	if c == nil {
		t.Fatalf("%v is not on the battlefield", id)
	}
	return c
}

// ebActivate gives me the energy, activates and resolves.
func EbActivate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, index, energy int, targets []game.TargetRef) {
	t.Helper()
	SetEnergy(t, g, me, energy)
	P7Activate(t, g, me, src, index, game.ActivateAbilityParams{Targets: targets})
}
