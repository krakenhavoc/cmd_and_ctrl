package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// improvise_test.go — improvise (CR 702.126) and the grant of it: Kappa
// Cannoneer's printed improvise and Archway of Innovation's "the next
// spell you cast this turn has improvise" (#3654).

const (
	KappaCannoneerOracle = "7fffbe7f-ad78-4bb6-a611-c37e495c6ee0"
	ArchwayOracle        = "bfa20bc7-4626-4a52-87f4-6e2763cb8ed5"
)

func ImpArtifacts(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = PushTapCostPermanent(g, owner, "Trinket", "Artifact")
	}
	return ids
}

func ImpKappa(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Kappa Cannoneer", TypeLine: "Artifact Creature — Turtle Warrior",
		OracleID: KappaCannoneerOracle, ManaCost: "{5}{U}", Colors: []string{"U"}, Power: 4, Toughness: 4,
		Owner: owner, Controller: owner}
}
