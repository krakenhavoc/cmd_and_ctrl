package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_collect_evidence_cards_test.go — #5193: the Murders at Karlov Manor
// cards built on collect evidence (CR 701.59), each played through its
// real catalog entry: the optional additional cost to cast, the activated
// ability's cost, and the effect form.

const (
	MkmCEPollenOracle     = "ec7d550b-9e13-4249-89c7-6ad43f71ea11"
	MkmCEMaskOracle       = "85ac3efc-5e19-4661-b52f-116f57e39581"
	MkmCEBiteOracle       = "ff387264-31da-4f6b-88c6-2843f56acbda"
	MkmCESpriteOracle     = "5e238e67-00f2-4656-bf69-0115e47321a5"
	MkmCEConfessOracle    = "96480e86-70a8-401d-abf8-a7d509311b73"
	MkmCEResearcherOracle = "75e7f7c7-fa98-497b-bf5e-8f74571e190c"
	MkmCEWhispererOracle  = "923b2a1b-d7b6-439d-9681-c7ba88165c31"
	MkmCEIzoniOracle      = "8074afbf-190b-4a24-b939-cc559bdffe38"
	MkmCEOrbOracle        = "bb039957-9e88-48cf-b41c-38d42eb87251"
	MkmCECollectorOracle  = "566b65bb-fa5d-4da6-a766-b08128ee0049"
	MkmCEInspectorOracle  = "22a33139-a730-4792-9dd5-2c0f99f96855"
)

// mkmCEEvidence fills p's graveyard with one sorcery per mana cost.
func MkmCEEvidence(p *game.Player, costs ...string) []uuid.UUID {
	out := make([]uuid.UUID, len(costs))
	for i, cost := range costs {
		out[i] = GyCard(p, "Evidence", "Sorcery", cost, "")
	}
	return out
}
