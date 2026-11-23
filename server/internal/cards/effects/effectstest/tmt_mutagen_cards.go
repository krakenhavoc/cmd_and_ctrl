package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_mutagen_cards_test.go — the Teenage Mutant Ninja Turtles cards that
// make a Mutagen token (#4676, #4457). Each test plays the real catalog
// entry and asserts what its text says.

const (
	TmtMuCommandoOracle   = "b24a87af-407f-4c58-80b4-caab9c65a233"
	TmtMuGenghisOracle    = "1d6d0348-46ad-46ed-ae7b-d7510f0ebe70"
	TmtMuBFFOracle        = "fe1eb8b3-c605-4711-a057-ed3565bfe75c"
	TmtMuWeirdnessOracle  = "170ab932-9d50-4ce7-9a42-08e1edce7e7c"
	TmtMuMutagenManOracle = "472d9a18-19a0-46c0-a4bf-328ebec7ce41"
	TmtMuChainOracle      = "ea9fc868-374b-4acc-be0d-4283907b4524"
	TmtMuOozeSpillOracle  = "c1c5f9da-1396-4406-b803-b64f66b8e49d"
	TmtMuRayFilletOracle  = "698098b8-b8e4-4e88-b13a-6b039579d192"
	TmtMuSewersOracle     = "a3ee8565-f34c-4073-bc94-3fe9ec6bbe8a"
	TmtMuCryptidOracle    = "f570bac8-9987-4963-af02-476d18abc847"
	TmtMuTheOozeOracle    = "82a6fa6f-13c2-4ea3-bec2-c51b56b2783b"
	TmtMuZooOracle        = "107e9ada-4e2e-4032-addd-274bca956621"
)

// tmtMuMutagens counts the Mutagen tokens `controller` controls.
func TmtMuMutagens(g *game.Game, controller uuid.UUID) int {
	return TmtB1Count(g, controller, "Mutagen")
}
