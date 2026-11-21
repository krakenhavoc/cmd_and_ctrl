package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch23_test.go — card-level coverage for the card-coverage
// roadmap's batch 23 (#385, `edhrec_rank` 2436–2535): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, draw or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b23-prefixed.

const (
	B23HordelingOutburstAlreadyOID = "a6450b8e-eb18-431c-9eb7-7daf107978b2"
	B23BorealDruidOracle           = "2fcc69ff-8ab5-4e14-afe3-db892049a872"
	B23DiscipleOfTheVaultOracle    = "c8625113-0ce4-4454-83a1-25c31b8bfb9a"
	B23PrimordialHydraOracle       = "1c36ed3a-c806-47e5-83f9-e44999c67fe5"
	B23IvyLaneDenizenOracle        = "a5da5ad6-4ed2-4041-a983-76a8c87fa109"
	B23EncroachingDragonstormOID   = "1e95c273-7fec-4f8d-8be9-e21ad4b93717"
	B23PyrohemiaOracle             = "9ac57a10-3402-4656-9079-f713884cde35"
	B23InsightEngineOracle         = "83409a22-10de-4d51-90a3-e0579ca8cbea"
	B23RegisaurAlphaOracle         = "0673f4e0-66ff-458c-b4ba-eb067e560cce"
	B23DionusOracle                = "d5e5cf55-eb9e-4f01-9056-791215411b27"
	B23ForcedFruitionOracle        = "448b27a5-7c0c-4ab6-bddf-bd62b920aacc"
	B23FangsOfKaloniaOracle        = "8d9bd4cc-564b-4fdf-9462-b4bb30583642"
	B23NivMizzetFiremindOracle     = "959acb66-84ca-4535-bca2-ad591895735e"
	B23ProteanHulkOracle           = "10180e2f-90c5-4d41-ba44-16b14948f923"
	B23RipApartOracle              = "cbdbf18f-0180-4ade-a79e-1e644dd42d6f"
	// #1210: both were declared skips — the activation gate and the
	// opponent-activation watch — and both landed with that issue.
	B23CollectorOupheOracle = "0c4bc9ea-a5fd-4f44-96a1-5448eee228c4"
	B23RunicArmasaurOracle  = "48e8ae59-a234-498e-9dae-bac8d1424ea5"
)

// b23TapMana activates a creature's first mana ability.
func B23TapMana(t *testing.T, g *game.Game, controller, card uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
}
