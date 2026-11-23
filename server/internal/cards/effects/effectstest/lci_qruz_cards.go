package effectstest

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_qruz_cards_test.go — The Lost Caverns of Ixalan, cards Q, R and U–Z
// (#4470). Each test plays the real catalog entry and asserts what its
// text says.

const (
	LciQRUZWhirlpoolOracle   = "43fb0f35-8eea-4344-9415-6ecb8aa46491"
	LciQRUZSpiketailOracle   = "41f6d6d9-3fe5-44b0-904d-1cb90d36f2b6"
	LciQRUZRayOfRuinOracle   = "e36b78ed-4e3e-406d-85b3-fe0deb493418"
	LciQRUZPrairieOracle     = "c071257a-63e7-48d0-a677-0b396a09b624"
	LciQRUZRidgelineOracle   = "4c0f4a63-586a-4dde-9621-b0dd9118b2e5"
	LciQRUZVentsOracle       = "696e7ddb-bdc7-40ee-bc5c-59e98f4a7401"
	LciQRUZHeraldGuideOracle = "b9687fef-4cb2-4bf1-91ad-49dada07f4d6"
	LciQRUZHeraldScoutOracle = "021428b1-005c-40ed-9d47-fb5aa970598b"
	LciQRUZLurkerBatOracle   = "e6f379ad-96b1-46c3-bbc6-a8cb57156a7f"
	LciQRUZRockslideOracle   = "0909cbce-b4a0-42bb-8a97-47eeb99606c0"
	LciQRUZBoulderOracle     = "57f15861-82cf-4faf-827c-9bc7f4844d2a"
	LciQRUZUnluckyDropOracle = "98ebde8c-b47d-440b-93cb-a1407b807074"
	LciQRUZInquisitorOracle  = "03ee49c9-95a3-44f1-b010-f7c4b2f86a7c"
	LciQRUZWanderglyphOracle = "ebc3ca0f-35eb-4606-90ab-eb64556f81fe"
)

// lciQRUZRestless puts a Restless land on the battlefield under the
// active seat and animates it with its row-0 ability.
func LciQRUZRestless(t *testing.T, g *game.Game, name, oracle string, generic int, colours ...string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	land := S58p6Push(g, me.ID, name, "Land", oracle, "", 0, 0)
	AdvanceTo(t, g, game.StepPrecombatMain)
	if c := S58p6View(t, g, land); c.IsCreature() {
		t.Fatalf("an idle %s is a creature", name)
	}
	NgheAnimate(t, g, me, land, generic, colours...)
	return land
}

// lciQRUZAssertAnimated checks a Restless land's animated characteristics.
func LciQRUZAssertAnimated(t *testing.T, g *game.Game, land uuid.UUID, subtype string, colours []string, power, toughness int) {
	t.Helper()
	c := S58p6View(t, g, land)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if !slices.Contains(c.Effective().Subtypes, subtype) {
		t.Errorf("subtypes %v, want %s", c.Effective().Subtypes, subtype)
	}
	if !slices.Equal(c.EffectiveColors(), colours) {
		t.Errorf("colours %v, want %v", c.EffectiveColors(), colours)
	}
	if c.CurrentPower() != power || c.CurrentToughness() != toughness {
		t.Errorf("%d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), power, toughness)
	}
}
