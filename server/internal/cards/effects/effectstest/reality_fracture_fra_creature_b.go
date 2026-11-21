package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_b_test.go — the Reality Fracture
// creatures of slice fra-creature-b.

const (
	RfcbDesperateFuturescribe = "83c12f50-e5a1-479d-afab-edb0b40412f2"
	RfcbDiviningDuelist       = "d0ce03fa-2284-4217-85b2-4db6b8a88c94"
	RfcbDraconicVisitor       = "6745e850-68f6-480b-a1e0-e06180dde169"
	RfcbEardrumRattler        = "25151d18-8e73-4b0d-a8d2-e5fa291e5729"
	RfcbEdgar                 = "1917bda6-c0e1-4c80-8009-74c28cf6b8e9"
	RfcbEmrakul               = "4421ab7d-6d9b-4edd-b5a0-53a8ed84da6f"
	RfcbFateshaperAspirant    = "bb3996b6-b68d-4ea1-b3dd-1600f8da2f1c"
	RfcbFblthpImpossiblyLost  = "17e7212e-69d8-4014-af98-447baed2cae7"
	RfcbFblthpKnowsTheWay     = "48b891b6-3175-4864-bb98-af131098c557"
	RfcbFlickeringHound       = "162421d2-8761-437f-bed9-578b61c96f1f"
	RfcbFrostbitePyromental   = "9a104f73-597b-49e3-8088-13db31f61900"
	RfcbGalliaTragicHost      = "58285b70-0d13-4723-b5a4-91fb6f9ccf03"
	RfcbGalliaMerrymaker      = "87f39199-3e4b-44fa-8406-62019eb43c10"
	RfcbGeistOfSaintThalia    = "ef32a4a9-14e2-4738-b4c2-53ce5e1d2a53"
	RfcbGhaltaImmovable       = "0d496c6e-8f7b-420f-9197-dacd2feb528f"
	RfcbGhaltaUnstoppable     = "8e5ea773-a60b-4502-9db5-ce93fa61cc89"
	RfcbGinger                = "5ea93956-085d-409b-bd95-dd669fa69eeb"
	RfcbGraftSurgeon          = "fe7938ce-6289-4bf8-a8f5-85ecccbb7f86"
	RfcbGreenhousePropagator  = "b6b77cb9-49a5-4d1a-ae2d-02ae176dc1fa"
	RfcbGrimRepriser          = "e853e2fb-f90d-4934-aa70-c8d3cdf12e56"
)

// rfcbPush puts a catalog creature on the active seat's side of the
// battlefield, not summoning-sick.
func RfcbPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}
