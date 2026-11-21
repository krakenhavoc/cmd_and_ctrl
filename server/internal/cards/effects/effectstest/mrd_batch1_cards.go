package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd_batch1_cards_test.go — Mirrodin, batch 1 (#5606, S82): the cards
// whose names start with A to C.

const (
	MrdAlphaMyrOracle          = "4caa8e68-e599-416f-b839-3d009569ff29"
	MrdAltarsLightOracle       = "fa9b6be2-b88c-4302-b7e2-faf25a60bcb9"
	MrdAssertAuthorityOracle   = "085bc7be-bd44-40af-8e8f-a1a8006fe22c"
	MrdAuriokBladewardenOracle = "2884e332-707c-4a97-adc1-e1317a513ff4"
	MrdAuriokTransfixerOracle  = "0ad5f4c5-f6d2-4e96-af65-5488d9711113"
	MrdBattlegrowthOracle      = "650bf82e-7f83-470c-bf4a-34281bfe9341"
	MrdBlinkmothWellOracle     = "dd3fd7cd-d8a1-48e2-861d-c7c40086dc75"
	MrdBoshOracle              = "2fb6f65f-a0fb-4d45-be1b-d6405e61d47a"
	MrdChimneyImpOracle        = "3901bf30-b7c1-4977-a7b1-fcdafcc266cd"
	MrdClockworkBeetleOracle   = "7b9ddff4-82f5-4068-8105-116b8a6180d5"
	MrdClockworkCondorOracle   = "3eff8f21-9a1b-4afc-8a7b-75a740d0a4ea"
	MrdClockworkDragonOracle   = "3361cf88-171a-4ddd-ac56-a31d42613c80"
	MrdClockworkVorracOracle   = "e1a64bba-eaff-4f59-b85d-7df0b13e2e6b"
	MrdCobaltGolemOracle       = "1e50af57-e49d-4b94-a370-44846d9f6b33"
)

// mrdPermanent puts a plain permanent on the battlefield.
func MrdPermanent(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// mrdEnter puts a catalog card onto the battlefield from its owner's hand,
// so its entry replacements apply, and lets it act this turn.
func MrdEnter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := Pr8Enter(t, g, owner, name, typeLine, oracle)
	FindBattlefieldCardForTest(g, id).SummonedThisTurn = false
	return id
}
