package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leg_b1_cards_test.go — Legends, batch 1 (#5639). Each test plays the
// real catalog entry and asserts what its text says.

const (
	LegB1PalladiaOracle  = "d7121ac7-e425-46cd-b006-1af391d97f87"
	LegB1VaevictisOracle = "4d2cc2a8-e08a-420a-8922-c63e39129e23"
	LegB1BolasOracle     = "e26bf6a9-b31c-4bc0-b55b-c01f2f69be6b"
	LegB1XiraOracle      = "4a6e367c-7bc9-44a3-8ede-f2d0651abad3"
	LegB1AdunOracle      = "aefd49a4-ecec-42cb-b882-e5db7f81eff8"
	LegB1TuknirOracle    = "0b6fa658-a344-4202-866b-e19997e32ff2"
	LegB1TorWaukiOracle  = "9439dfa0-daf8-4dac-af32-755fc2984126"
	LegB1CaleriaOracle   = "74284bbc-b011-4e16-b8d5-7ddacc77f9d7"
	LegB1RamsesOracle    = "6c9a9071-8159-4607-9465-de796f1dc2cc"
	LegB1JacquesOracle   = "c030ca14-33cb-40b3-a1f5-b6d0cb0efd49"
	LegB1KarakasOracle   = "59119143-c0fa-49dd-adf0-e2fd3029c48b"

	LegB1ElderDragonType = "Legendary Creature — Elder Dragon"
)

// legB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func LegB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}
