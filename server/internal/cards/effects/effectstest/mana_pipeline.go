package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_pipeline_test.go — the S32 mana-pipeline batch (#352): the ten
// Signets, Cabal Coffers, Gaea's Cradle, Temple of the False God, Mox
// Opal, Exotic Orchard, Reflecting Pool, Mox Amber, Delighted
// Halfling, Shrine of the Forsaken Gods, Eldrazi Temple, and the
// Fellwar Stone retrofit.
//
// Four seams, and each card here is chosen to prove one of them
// against the printed text rather than against the implementation:
//
//   - a mana component in a mana ability's cost (Signets, Coffers);
//   - an activation gate (Temple, Mox Opal, Shrine);
//   - a produced string computed at activation, derived (Orchard,
//     Pool, Fellwar, Amber) or scaled (Coffers, Cradle);
//   - spend restrictions that actually bind at SPEND time (Halfling,
//     Shrine, Eldrazi Temple).
//
// The last group carries the weight. A restriction stamped but not
// enforced makes every one of those cards stronger than printed,
// which is the #259 rule, so each has a test that pays for something
// it should NOT be able to pay for.

const (
	AzoriusSignetOracle    = "e018773f-95b3-49a3-9674-6f04ddef2092"
	DimirSignetOracle      = "7d881c57-0bd9-4c57-aa4a-b10808b86143"
	SelesnyaSignetOracle   = "1436dd81-496e-42a5-b210-fb5b9cdf073f"
	CabalCoffersOracle     = "7358e164-5704-4e78-9b21-6a9bf2a968ce"
	GaeasCradleOracle      = "7c427c3d-ecd8-45ef-bebd-8f10f4a311db"
	TempleFalseGodOracle   = "cfdd5dc6-593e-495a-8cfe-3a56b3c4c7df"
	MoxOpalOracle          = "de2440de-e948-4811-903c-0bbe376ff64d"
	MoxAmberOracle         = "7a43bd27-fdd8-41f0-9bc4-92568f3408f1"
	ExoticOrchardOracle    = "27b047e3-0d41-45e2-98e9-9391d7923a1e"
	ReflectingPoolOracle   = "67f43ac6-2a58-4b53-b5d7-0330e2a252e2"
	DelightedHalflingOracl = "f9d3b046-0b95-4103-a630-4b3fb88bb60b"
	ShrineForsakenOracle   = "8ea46945-d5ab-4209-b473-4769e7b8b962"
	EldraziTempleOracle    = "7fab8d65-af51-47d3-8f10-2676bf6e8ba3"
)

// seedManaLand puts a land on the battlefield with an explicit
// ProducedMana list, which is how the deck importer stamps Scryfall's
// `produced_mana` and therefore what the derivation reads for lands
// outside the catalog.
func SeedManaLand(g *game.Game, owner uuid.UUID, name, typeLine string, produces ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:   id,
		Name:         name,
		TypeLine:     typeLine,
		ProducedMana: produces,
		Owner:        owner,
		Controller:   owner,
	})
	return id
}

// manaPickFor returns the latest queued colour pick for a chooser.
func ManaPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == chooser {
			out = c
		}
	}
	return out
}
