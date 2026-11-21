package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// voting_a_test.go — #2928, the first slice of the other voting cards on
// ADR 0146's seam: Tyrant's Choice, Bite of the Black Rose, Capital
// Punishment, Lieutenants of the Guard, Orchard Elemental, Messenger
// Jays, Emissary Green, Travel Through Caradhras, Fateful Tempest, Sail
// into the West, Split Decision, Council Guardian.

const (
	TyrantsChoiceOracle           = "ee7d2c2f-e32f-4be6-82ac-ff1e32055376"
	BiteOfTheBlackRoseOracle      = "4db04e23-e1da-4c41-9dba-5502dfff73fb"
	CapitalPunishmentOracle       = "c2902f4f-b40e-49f7-8ab0-5b6a70021482"
	LieutenantsOfTheGuardOracle   = "80745dea-2335-43fc-9915-f62e0436e73b"
	OrchardElementalOracle        = "189b7c2b-81df-41c3-94ae-6b32dbd48cec"
	MessengerJaysOracle           = "dfeb1d36-d1d9-43dd-a79f-a98560676bb3"
	EmissaryGreenOracle           = "c43e7fa5-ff93-42d9-af79-ee417a8e41a7"
	TravelThroughCaradhrasOracle  = "2d47a98b-5700-4096-b080-32df4d79910c"
	FatefulTempestOracle          = "9d2f02f7-adcd-4d23-a051-ebca70f22ba5"
	SailIntoTheWestOracle         = "0a8956d0-bd37-4305-8443-4df043796f6e"
	SplitDecisionOracle           = "5f725d51-e133-488c-9ca9-1c7b0523b5bd"
	CouncilGuardianOracle         = "4f212675-e34d-4387-a125-e17b25757083"
	VotingATestCreatureTypeLine   = "Creature — Bear"
	VotingATestPlusOneCounterKind = "+1/+1"
)

// voteRound casts every seat's ballot, in turn order from the active
// player, for the labels given in that order.
func VoteRound(t *testing.T, g *game.Game, labels ...string) {
	t.Helper()
	order := VoteOrder(g)
	if len(labels) != len(order) {
		t.Fatalf("voteRound: %d labels for %d seats", len(labels), len(order))
	}
	for i, p := range order {
		VoteFor(t, g, p.ID, labels[i])
	}
}
