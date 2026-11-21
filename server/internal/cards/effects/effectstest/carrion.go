package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// carrion_test.go — Carrion Rats and Carrion Wurm (ADR 0187 §3 and §4,
// #3571): an attacks-or-blocks trigger whose any-player offer is paid by
// exiling cards from the payer's own graveyard, and whose "if a player
// does" makes the creature assign no combat damage this turn.

const (
	CarrionRatsOracle = "1aa6b31b-cd30-4741-abc6-2ef72e445630"
	CarrionWurmOracle = "ed0cf504-c365-4486-92bc-c329b83b99d0"
)

func CarrionPush(g *game.Game, owner *game.Player, name, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: "Creature — Test",
		Power: power, Toughness: toughness, Owner: owner.ID, Controller: owner.ID,
	})
}

// carrionToOffer passes priority until the trigger resolves and its
// offer opens.
func CarrionToOffer(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8 && AppOpenLink(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if AppOpenLink(g) == nil {
		t.Fatal("the trigger never opened an offer")
	}
}

// carrionAttack has `atk` attack `at` and brings the trigger's offer up.
func CarrionAttack(t *testing.T, g *game.Game, atk uuid.UUID, at *game.Player) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, at.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	LockInAttacks(t, g)
	CarrionToOffer(t, g)
}
