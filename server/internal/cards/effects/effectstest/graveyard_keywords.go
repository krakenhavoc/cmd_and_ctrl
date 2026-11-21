package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_keywords_test.go — the catalog half of #1221: unearth
// (CR 702.84), scavenge (CR 702.97), embalm (CR 702.128) and
// eternalize (CR 702.129). The engine half — the zone dimension on
// the activation path, the ExileSelf cost component, the enumerator's
// walk — is in game/other_zone_ability_test.go and
// legal/other_zone_ability_test.go.

const (
	DregscapeZombieOracle = "be9d1346-4416-4ade-ae84-7a4121e0bd12"
	DeadbridgeGoliathOrac = "1498f5a1-6df7-4f80-9470-c93528b64a9c"
	SacredCatOracle       = "d85ea576-a794-44bf-b405-1f1c49477409"
)

// pushCatalogGraveyardCard seeds a catalog card into a player's
// graveyard — the zone every keyword in this file activates from.
func PushCatalogGraveyardCard(p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness,
		Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// activateFromGraveyard walks the cursor to the active seat's own
// precombat main (every keyword here is sorcery-speed), seeds the
// card into that seat's graveyard, floats `mana` and activates
// ability 0.
func ActivateFromGraveyard(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, mana string, params game.ActivateAbilityParams) (uuid.UUID, *game.Player) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := PushCatalogGraveyardCard(me, name, typeLine, oracle, power, toughness)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, params); err != nil {
		t.Fatalf("%s: activate from graveyard: %v", name, err)
	}
	return id, me
}
