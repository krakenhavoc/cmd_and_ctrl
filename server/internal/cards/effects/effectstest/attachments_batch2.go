package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_batch2_test.go covers the S24 tail: six more Equipment
// and two more Auras on the surface #379 built, plus the CR 704.5m
// branch that shipped with them (an Aura attached to NOTHING goes to
// its owner's graveyard) seen through a real catalogued card rather
// than through a fixture.
//
// The shared helpers — advanceToMain, equipTo, attachmentHostOf,
// effectiveAbilities, dealCombatDamageToPlayer — all live in
// attachments_test.go and its siblings.

const (
	WarhammerOracle      = "dba35ac5-7ad3-488a-a006-6b9a1d54eea5"
	BasiliskCollarOracle = "f5f4dd28-f4ae-4d39-b9b8-6ebfd63c93fe"
	DarksteelPlateOracle = "b5b4cf54-ed5e-42d0-9d98-5fec76b0b0b8"
	MaskOfMemoryOracle   = "d6b2c998-a226-426c-a40d-6e6007041bfe"
	WhispersilkOracle    = "9ad4f730-a18e-4a7c-a468-a926c718c741"
	ColossusHammerOracle = "8ec03b88-8d3a-4a32-8b7c-7da59b0c03d0"
	BattleMasteryOracle  = "a06c7c1a-8534-40fb-8bb8-59f8f2530567"
	ControlMagicOracle   = "cd0d7141-46d2-4aa3-bc77-6b3b4513803e"
)

// seedBear puts a vanilla 2/2 on the battlefield under `owner`.
func SeedBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// seedEquipment puts an Equipment on the battlefield unattached,
// which is where every Equipment starts.
func SeedEquipment(g *game.Game, owner uuid.UUID, name, oracle string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: EquipTypeLine,
		OracleID: oracle, Owner: owner, Controller: owner,
	})
}
