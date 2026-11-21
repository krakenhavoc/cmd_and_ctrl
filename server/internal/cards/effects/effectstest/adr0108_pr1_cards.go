package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_cards_test.go — ADR 0108 Delivery PR 1 (#1886, #1887):
// "if it would die this turn, exile it instead" and "can't be
// regenerated this turn", each card against the board that tells its
// wording apart.

const (
	P1LavaCoilOracle      = "fa71db44-5181-4c51-8b24-7fbedf36e3ca"
	P1DemonfireOracle     = "314a5c76-1a68-433b-a383-1834400254a8"
	P1DisintegrateOracle  = "92d6af2f-728e-4e41-87cb-5c90878a2f2f"
	P1IncinerateOracle    = "d8fd7a34-8418-4e98-b79b-119c4348c667"
	P1AngerOracle         = "3a7fe095-8278-4b1d-bec4-19b35bdcdd1b"
	P1FlayingOracle       = "cc2d016a-af44-427b-a25a-593274369449"
	P1EclipseOracle       = "5beb8d6e-d3c1-46a5-8516-d6bf66413cff"
	P1MalfunctionOracle   = "c7ecaa1a-fbf7-436b-a1c9-d7810b0dc5dc"
	P1WhippoorwillOracle  = "84050a10-e1f1-413e-aa21-5c1f47bb2a64"
	p1PreventEverything   = 20
	P1BearPower, P1BearTo = 2, 2
)

// p1Creature puts a vanilla creature on the battlefield for `owner`.
func P1Creature(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Creature", TypeLine: "Creature — Bear",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func P1Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

func P1Regenerate(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(id); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
	})
}

func P1ShieldCreature(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { g.PreventNextDamageThisTurnForEffect(uuid.Nil, id, p1PreventEverything, false, "shield") })
}

func P1WantZone(t *testing.T, g *game.Game, id uuid.UUID, want game.ZoneKind, what string) {
	t.Helper()
	if z := ZoneOf(g, id); z != want {
		t.Errorf("%s is in %q, want %q", what, z, want)
	}
}
