package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_test.go — card-level coverage for the card-coverage
// roadmap's batch 02 (#295, `edhrec_rank` 238–360): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation, land play or attack. Every batch-specific
// name is b02-prefixed — three batches are being written in parallel.

const (
	B02UndergroundSeaOracle      = "4b22be3a-8ce1-47d1-b82e-6c3ccfb0548b"
	B02VolcanicIslandOracle      = "c718911c-c955-4eb9-9e16-be4bd49a4e4e"
	B02TropicalIslandOracle      = "74b7fe23-5d3a-4092-8d78-7c0eba8f6f73"
	B02TundraOracle              = "02418479-9455-417f-a6a1-004356faff37"
	B02SundownPassOracle         = "5ad0b405-cca4-475e-985c-4d7e3599d87e"
	B02HauntedRidgeOracle        = "e2a37967-4212-4553-9f77-bcb613405807"
	B02OvergrownFarmlandOracle   = "709d2f10-1585-48c3-9058-ddd5f62f0452"
	B02DeathcapGladeOracle       = "f6d24565-5b32-4eff-b2e0-6e2c25516ff0"
	B02ArcaneSanctumOracle       = "7d7cf15c-06b9-4062-a1eb-32614c458a3b"
	B02JungleShrineOracle        = "2e69537c-c898-4e13-a72d-ce3957a90304"
	B02SimicGrowthChamberOracle  = "046f5783-cc7b-416a-8cf6-2bcef9c2cc1a"
	B02GolgariRotFarmOracle      = "1b301478-b14f-4ef8-94e6-9647d582eabe"
	B02SeatOfTheSynodOracle      = "39451b4d-cd7a-40da-b457-cb51b609173f"
	B02ScavengerGroundsOracle    = "5ece7d03-9ee7-4953-a06e-9d8e41874903"
	B02InfernalGraspOracle       = "94f0a572-e91c-4b56-a5d1-6cbbeabd210d"
	B02WitheringTormentOracle    = "ffce81c5-1b58-4882-a4e7-6f8d7cb170de"
	B02DamnOracle                = "b01d61cc-9844-4191-86a0-f2db6d42d6e5"
	B02SnapOracle                = "ac914d98-221e-426c-8a50-342896b15f9e"
	B02ArchmageEmeritusOracle    = "8305d576-21d8-4ce7-8eda-a7cd9793aca5"
	B02BalefulStrixOracle        = "37688720-03de-4eca-a82d-a0afe8d58adc"
	B02GuardianProjectOracle     = "4f9e07ae-6341-4b46-9f77-f17ab659d266"
	B02LotusCobraOracle          = "8ad91f64-ccab-4edc-bd54-b2ee9267d614"
	B02AvengerOfZendikarOracle   = "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7"
	B02LoranOfTheThirdPathOracle = "b3d81980-76f2-44e2-b1c9-01e30c726312"
	B02SyrKonradOracle           = "14c3ff84-1e82-4606-a433-869fc52cc382"
)

// b02TopOfLibrary puts a card with the given type line on top of a
// player's library (PushTop is the top: the last element).
func B02TopOfLibrary(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}
