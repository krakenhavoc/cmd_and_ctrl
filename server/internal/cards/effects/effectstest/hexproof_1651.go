package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hexproof_1651_test.go — #1651's proof cards (ADR 0038's amendment of
// 2026-09-28): Detection Tower's turn-scoped hexproof waiver (B1), and
// "loses <keyword> and can't have <keyword>" in its scoped (Arcane
// Lighthouse) and static (the Archetype cycle) forms (B2). The engine
// half is in game/cant_have_1651_test.go.

const (
	DetectionTowerOracle         = "93695c16-c441-492d-af12-b57df9739846"
	ArcaneLighthouseOracle       = "30ac68e6-160a-41f9-9f0f-0e0eef383150"
	ArchetypeOfEnduranceOracle   = "79254223-3fa6-4b7c-8163-1e48bf5cb708"
	ArchetypeOfImaginationOracle = "a5458de0-0f61-49a3-a013-d90f92559809"
	ArchetypeOfFinalityOracle    = "eb4e9ce9-1917-46b1-b01c-645f5920ef9c"
)

// activateLand puts the land under the active player, floats {C} for
// the {1}, and activates its first non-mana ability, resolving it.
func ActivateLand(t *testing.T, g *game.Game, name, oracle string) (me *game.Player, land uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	land = PushStampedCatalogCard(g, me.ID, name, "Land", oracle)
	AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return me, land
}

// pushStampedCatalogCard seeds a catalog permanent through the entry
// event, so it carries a battlefield timestamp (CR 613.7d) and the
// layer version moves.
func PushStampedCatalogCard(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}
