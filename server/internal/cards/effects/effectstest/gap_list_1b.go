package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap_list_1b_test.go — the second slice of tracker #293's "Gap list
// 1/2": one or more tests per card.

const (
	Gl1bPurphorosOracle       = "4fdbbec2-e921-4b63-958d-f9ba1e417197"
	Gl1bCrucibleOracle        = "33c722cf-b4bf-431f-aefd-ee96241a7fbf"
	Gl1bSculptingSteelOracle  = "6c85271c-c711-49b0-a72e-9e576c33714d"
	Gl1bDrannithOracle        = "aadd10d0-6dd0-4bdc-8d93-ff08e29a5863"
	Gl1bFerrousLakeOracle     = "62c15af0-40e1-407d-b056-7a3d909e3fdb"
	Gl1bOverflowingBasin      = "5ac8e01c-b0a7-4855-a122-1cd26b07c4a5"
	Gl1bElegantParlorOracle   = "9ea747cf-5d04-4aa7-bdc3-8145860cd1ba"
	Gl1bThornspireVergeOracle = "e861bc08-4f0b-4d22-9b85-9d20227fd5b4"
	Gl1bBlackMarketOracle     = "21338b71-37f5-4121-9b84-565025ebcd17"
	Gl1bGreenwardenOracle     = "3bcf090c-e890-4a9f-a8aa-6079e4ec9947"
	Gl1bProdigyOracle         = "2e2ace5b-4018-43af-8e72-ebafec1a7739"
	Gl1bSharedAnimosityOracle = "a27445db-33f2-4571-98b5-83206b797484"
	Gl1bHellkiteTyrantOracle  = "d9b066ff-9519-415c-ae17-bfea703c9889"
	Gl1bLoyalApprenticeOracle = "7f268b15-ac92-4e98-821b-78d15d9285d9"
	Gl1bSafekeeperOracle      = "bddf8f4a-3149-4dd6-a9e5-7747e7e45a1c"
	Gl1bRuinCrabOracle        = "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7"
)

func Gl1bNamed(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

// --- Harmonic Prodigy ------------------------------------------------------------

func Gl1bPyromancerTokens(g *game.Game, owner uuid.UUID) int {
	return len(Gl1bNamed(g, owner, "Elemental"))
}
