package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b14_cards_test.go — Avatar: The Last Airbender, batch 14 (#4459).
// Each test plays the real catalog entry through the engine.

const (
	AvB14CubOracle        = "2b0afb89-0944-4861-b9c3-e909e2ac215e"
	AvB14BoulderOracle    = "d33ee141-a880-4368-a274-a6ae2f2c624b"
	AvB14BumiOracle       = "f7493bcf-5c17-4562-a759-37082ec49ff8"
	AvB14ZookeeperOracle  = "6a84a8a4-c8b0-45ce-af6f-605fcffc395b"
	AvB14GeneralOracle    = "e43c31ee-1bac-4f96-8b74-334b3238ab73"
	AvB14RumbleOracle     = "f04a7763-5fba-42e5-a45a-b3dc2daba902"
	AvB14AscensionOracle  = "f8966fd5-2926-4ba0-b7d6-cfea40cd4e05"
	AvB14FlopsieOracle    = "a8966245-501c-47e4-8ce0-849dcb532eb2"
	AvB14VinebenderOracle = "82fefb82-a436-488d-8ecd-594686117e34"
	AvB14HaruOracle       = "7488ce54-37c6-475c-b327-59f52f2b1424"
)

// avB14Forest parks an untapped Forest under owner.
func AvB14Forest(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushEarthbendLand(g, owner, "Forest", "Basic Land — Forest")
}
