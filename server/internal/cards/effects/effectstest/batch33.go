package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch33_test.go — card-level coverage for the card-coverage
// roadmap's batch 33 (#396, `edhrec_rank` 3447–3548): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	B33VisionsOfBeyondOracle     = "c68964df-51ae-43b0-abea-74735016c13f"
	B33AlelaOracle               = "1cae5752-b4af-4a8f-8c8c-2493e163083b"
	B33CoalitionRelicOracle      = "008cb342-79f5-4df6-a6b7-0e9e22ed693f"
	B33BurningInquiryOracle      = "af98156b-3064-4f16-940e-10039241f2b0"
	B33LathielOracle             = "d3c56fc4-3611-41b1-952e-4c5311b1510b"
	B33RaidBombardmentOracle     = "1734e777-d34a-4526-ae24-ef034f147cc5"
	B33BlastingStationOracle     = "3a38d2d1-c4ff-4088-b1df-5feb9602ee2e"
	B33MajaOracle                = "0a2075b5-9609-433d-bcd2-e0a637456cf8"
	B33DalkovanEncampmentOracle  = "33a90122-7280-4481-9b97-5879194cae40"
	B33CountersquallOracle       = "6df620b8-1e54-4f63-8556-12c75e5679af"
	B33NestOfScarabsOracle       = "1f21cf59-6390-44b4-ab2e-ef290dfd8c85"
	B33RecklessBarbarianOracle   = "a621ae15-122a-4e11-bfcf-11dfc09784c0"
	B33CloudblazerOracle         = "f84d1291-1f82-4b67-a26e-b79624b4ce1d"
	B33RakdosJoinsUpOracle       = "6a47865c-8fa2-4cb2-aebf-8009c065395f"
	B33GoodFortuneUnicornOracle  = "1e0cdff3-3ec5-41fc-8053-f072bae156b3"
	B33OverseerOfTheDamnedOracle = "c1f085ce-5b52-45b3-aef0-f77f36b3da36"
	B33PatronOfTheVeinOracle     = "8dbc8fdb-36ce-4f30-b679-9c2029fcd9c6"
	B33FervorOracle              = "8e0cea9c-3110-4728-9378-76849e33bb90"
	B33HermesOracle              = "63e2cba7-ec1a-44bf-b915-d2ae75851430"
	B33SidisiOracle              = "3fad7072-21e3-446e-a28f-615038c8bfea"
	B33FlowOfKnowledgeOracle     = "b65f20de-52fa-4904-bbfe-7ba53ccd8ae5"
	B33CopperMyrOracle           = "8b52f30c-5e38-4333-88ab-901b37105b36"
	B33PlanarEngineeringOracle   = "dc48079a-b1c4-4c43-a989-418c78a264ec"
	B33DoublingCubeOracle        = "9afd8f12-0796-4500-aaa3-10b4a46ef6ec"
	B33StoneOfErechOracle        = "73dad679-1edb-41c9-9d43-56dc93c3e9fe"
	B33GodEternalBontuOracle     = "183891b0-b5ec-47f4-8d09-b9d3cfc4e7f1"

	// Contagion Clasp puts a -1/-1 counter on target creature as it
	// enters — the batch's -1/-1 counter source for Nest of Scarabs.
	B33ContagionClaspOracle = "43f2d81e-aa01-4fa9-9046-6a27a05dbd2d"
)

// b33Goaded reads a battlefield card's goad marker.
func B33Goaded(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.LatestGoader()
}

// b33AttackingTokensNamed lists the tokens named `name` under
// `controller` that are attacking `defender`.
func B33AttackingTokensNamed(g *game.Game, controller, defender uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.AttackingTarget == defender {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
