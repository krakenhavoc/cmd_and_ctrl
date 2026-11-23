package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lose_land_types_cards_test.go is ADR 0109 Delivery PR 3 (#1604):
// "loses all land types" (CR 205.3i, 613.1d), from a resolved effect
// (Ultima, Origin of Oblivion's blight) and from a static (Lithoform
// Blight, Alpine Moon).

const (
	LlUltimaOracle          = "baa337ce-edc6-4ee5-a898-68e9dbb4ab93"
	LlLithoformBlightOracle = "b256c311-e37d-4b3c-889c-85443e3d5e7e"
	LlAlpineMoonOracle      = "8b46b50c-f824-4c4e-86de-38065c6f9a64"
)

// llAttackWithUltima seeds Ultima for `me`, attacks `opp` with it, aims
// the blight at `land` and resolves the trigger.
func LlAttackWithUltima(t *testing.T, g *game.Game, me, opp, land uuid.UUID) uuid.UUID {
	t.Helper()
	ultima := PushCatalogPermanent(g, me, "Ultima, Origin of Oblivion", "Legendary Creature — God", LlUltimaOracle, false)
	DeclareAttack(t, g, opp, ultima)
	if OpenPickTarget(g) {
		AnswerPickTarget(t, g, land)
	}
	item := TriggerOnStack(g, ultima)
	if item == nil {
		t.Fatal("Ultima's attack trigger is not on the stack")
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != land {
		t.Fatalf("the trigger targets %v, want the land %s", item.Targets, land)
	}
	PassPriorityAroundTable(t, g)
	return ultima
}
