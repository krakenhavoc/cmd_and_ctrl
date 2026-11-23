package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_conditions_test.go — #743, the catalog half of ADR 0020's
// activation-condition addendum: the condition builders in
// activation_conditions.go and every card that declares one. The
// engine half (nothing paid, no resolution re-check, the error order
// against sorcery speed) is in game/activation_condition_test.go.
//
// Each card test runs the ability with its condition false and asserts
// the refusal paid nothing, then makes the condition true and asserts
// the printed effect.

const (
	AcTectonicEdgeOracle       = "4927150d-7ff6-4232-b20e-d2ea245ac710"
	AcWeatheredWayfarerOracle  = "1093fdb6-bd4f-46a8-92e3-5aa46b52bb4f"
	AcBondersEnclaveOracle     = "f33ce38a-34ec-4b65-a0fc-160484a02007"
	AcCrypticCavesOracle       = "b2eb7a64-a307-4a78-a25d-63fb3ae1e237"
	AcEndlessAtlasOracle       = "cb923c4d-1d2e-4039-90bb-72d1981b9738"
	AcKelpieGuideOracle        = "8fc1e8b1-3afc-4e9d-a3ae-7bd9bcdcb465"
	AcBarbarianRingOracle      = "eeb9377b-72c1-4214-9a66-0f55577c17d1"
	AcCephalidColiseumOracle   = "c733873e-77db-471f-8061-139db24f7e7c"
	AcRivendellOracle          = "2550099d-b3e2-4eb6-9f36-0fc412828ca6"
	AcSeaGateWreckageOracle    = "91f34686-cb96-49c0-b4a7-49dd1fd076e2"
	AcMinasTirithOracle        = "7b0d7e62-0287-454a-8702-b0bfa7b41245"
	AcIdolOfOblivionOracle     = "c232d974-9c65-4bce-b045-e5a7cc0c62e0"
	AcSpeakerOfTheHeavens      = "e4bacf1b-0a45-4c73-857d-dffadb7e6fa5"
	AcLuminarchAscensionOracle = "90076bf5-aa9a-4a6e-9035-9aa97fd5561e"
	AcSliverHiveOracle         = "e7286688-ffbe-4d25-ad55-27990f005368"
	AcInventorsFairOracle      = "91d4a5fe-fd6d-4b14-a63f-61b4d0ecd9c4"
	AcLeechriddenSwampOracle   = "d83c86c1-126d-49e9-9b13-9e55784c49c5"
	AcMistveilPlainsOracle     = "bb5c1817-ac22-4779-9005-251bc354f181"
	AcLagomosOracle            = "12c98c1e-7e02-473d-b494-ad8a0731fe82"
	AcHallOfOraclesOracle      = "0bb896ba-e15e-43a9-9120-e674d7ba003c"
	AcSeedcoreOracle           = "249fdd3e-376c-4ec2-a612-4353e0e61ee2"
	AcGoroGoroOracle           = "b80df711-af51-4f9a-9a3f-f7a4cddc9c2c"
	AcLilypadVillageOracle     = "5bb06e6f-e3af-4caa-b66d-77248ad46b61"
	AcArchOfOrazcaOracle       = "3bb518ff-399b-4ce7-b9ad-a1d563dd7792"
)

// --- helpers ---------------------------------------------------------

// acRefused activates and requires ErrConditionNotMet with the source
// untapped and the stack untouched.
func AcRefused(t *testing.T, g *game.Game, player, card uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	stackBefore, metaBefore := g.Stack.Size(), len(g.StackMeta)
	err := g.ActivateCatalogAbility(player, card, idx, params)
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("ability %d: err = %v, want ErrConditionNotMet", idx, err)
	}
	if effects.OnBattlefield(g, card) && B16Tapped(t, g, card) {
		t.Error("a refused activation tapped the source")
	}
	if g.Stack.Size() != stackBefore || len(g.StackMeta) != metaBefore {
		t.Error("a refused activation put something on the stack")
	}
}

// acLands seeds n vanilla lands under owner.
func AcLands(g *game.Game, owner uuid.UUID, n int, name string) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = SeedManaLand(g, owner, name, "Land", "C")
	}
	return out
}

// acPermanent seeds a permanent with the given type line and colours,
// able to attack and tap.
func AcPermanent(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// acPlayerRef is a one-player target list.
func AcPlayerRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func AcGraveyardCards(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		PushGraveyardCardForTest(p, "Graveyard Filler")
	}
}
