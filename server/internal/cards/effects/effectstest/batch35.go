package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch35_test.go — card-level coverage for the card-coverage
// roadmap's batch 35 (#398, `edhrec_rank` 3650–3752): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	B35SphinxsTutelageOracle    = "158db438-52ba-4c30-b7bf-d1c4f76da917"
	B35RiskyShortcutOracle      = "a7c55495-6aca-4d17-b61a-cab8fa38b9f4"
	B35SealOfPrimordiumOracle   = "f14dbb39-c9f9-4f64-b22a-38dba28f5b1e"
	B35ScaledNurturerOracle     = "13b96709-0e88-476b-9485-956e682bb818"
	B35AetherGaleOracle         = "b21d6482-1b3a-47e6-98a5-3067f5f3818b"
	B35PlagueBelcherOracle      = "f086a82d-8dbe-4d02-a076-03789704e9d4"
	B35MemoryErosionOracle      = "f4a96881-586d-44ad-b427-5cdf8988f9a1"
	B35SmileAtDeathOracle       = "9228f04c-506f-4453-a335-e66876b8ce8d"
	B35LordOfExtinctionOracle   = "ea5e3401-bd6c-47bb-a52a-8eec5f09455d"
	B35TheNecrobloomOracle      = "b981af39-4ee6-4fbc-9a89-618dcad9dfbf"
	B35GarnaOracle              = "3fb83218-e18d-405d-91c8-46b5e4e672a9"
	B35DreadSummonsOracle       = "642fa01d-025e-44dd-8360-5325e5a28282"
	B35AbandonAttachmentsOracle = "82333385-631f-4abf-b159-bb367f1c6fd9"
	B35LuxArtilleryOracle       = "cbd76b22-d04e-48e7-bcab-c7f5466d67d8"
	B35AbsorbOracle             = "132ca99a-a3c7-4ed6-b4d0-0edcd7140ca2"
	B35EdenOracle               = "84856b92-5ce8-47f3-9a1c-78d6a3e26aca"
	B35IkraShidiqiOracle        = "a1a1761c-88e5-40b4-ba4c-60735c054b09"
	B35ScreamingNemesisOracle   = "fcb7c93c-46ab-49b5-a6e0-35d73f3be8f0"
	B35WildwoodScourgeOracle    = "b9dec104-c636-4770-a7fc-7a3331face15"
	B35NekusarOracle            = "8a5e3c8e-8e22-49b9-8ee5-4a36361f0da6"
	B35ThirstForDiscoveryOracle = "1e05e6ef-14af-451d-9d54-e75b1f8871ab"
	B35FinalActOracle           = "335a6c6a-030f-45d4-806c-467a22962eed"
	B35GreenwardenOracle        = "2facb1b3-4522-4610-a0ab-29ac53ca7fcd"

	// Deadly Tempest is the batch's 28th ready card and was already on
	// main from the boardwipe primitives (#382); pinned so the table
	// still accounts for every card on the issue's list.
	B35DeadlyTempestAlreadyOnMainOracle = "b5516bc9-ec8d-4323-8748-96c49d7d0622"
)

// b35Block declares `blocker` as a blocker of `attacker` in the
// declare-blockers step.
func B35Block(t *testing.T, g *game.Game, blocker, attacker uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
}
