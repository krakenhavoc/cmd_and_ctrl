package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch30_test.go — card-level coverage for the card-coverage
// roadmap's batch 30 (#393, `edhrec_rank` 3145–3244): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, block, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b30-prefixed.

const (
	B30StonespeakerCrystalOracle = "a494fcee-6885-434c-aad5-6f83640c4472"
	B30LunarConvocationOracle    = "cb68d1e4-36aa-4a00-b671-8959e7d526d4"
	B30TolarianWindsOracle       = "686cce6e-18ec-45b0-8d2e-74fa353a905e"
	B30WoodlandChasmOracle       = "bf5482b6-dd3e-4fb7-bc62-29e23b417a5f"
	B30RiversRebukeAlreadyOracle = "c52cfb41-18f3-4e73-b5e7-d75baf74e578"
	B30DamnablePactOracle        = "28dd0ae4-fc56-4abf-af24-6c6c8d0a10cd"
	B30HonoredDreyleaderOracle   = "44bffa3e-4d31-4f9b-829c-e2066ab2f641"
	B30GoblinTrashmasterOracle   = "0283bf5e-ddf2-4a4a-a7cf-d3e27eed7e7d"
	B30KhalniGardenOracle        = "b2d5ba45-8674-4428-89db-c2bbbf0bf5c5"
	B30WindgracesJudgmentOracle  = "de1ca6ed-b275-4f62-ba05-f31b3659b352"
	B30ThrabenCharmOracle        = "7eccd10d-f224-4467-af25-f2b768993f9f"
	B30SurlyBadgersaurOracle     = "0209dc74-ac49-4deb-907a-e9fa49d27a0f"
	B30PromisingVeinOracle       = "861eb7d7-7616-4620-a4fd-4b8c3bf00dd1"
	B30OppressionOracle          = "f488f679-f5d5-4e61-b0c8-0b0eb622df0d"
	B30PickYourPoisonOracle      = "9af4a832-d634-47aa-91ed-79d44fe08864"
	B30KeeperOfFablesOracle      = "c8ca3116-e0f0-4e27-aa0f-99ed85927040"
	B30IronSpiderOracle          = "e123fd7d-ace9-48a4-9510-eedcc837d8e8"
	B30FavorableWindsOracle      = "2361ca87-6352-4ba3-8d91-b3d71242914d"
	B30CanopyTacticianOracle     = "8b20d6f4-6322-4435-92d5-acaae74774f4"
	B30VorelOracle               = "3001f971-c4db-4176-89b5-e7f4a8890c0f"
	B30HourOfPromiseOracle       = "51ff3e53-90bb-41bf-80e4-bb3fd51d574a"
	B30SatoruOracle              = "7555c429-5f2d-4171-b6b0-8e3c8da7f314"
	B30BetorOracle               = "826a0db6-195a-4697-9bd6-f544b279e3cf"
	B30WakeningSunsAvatarOracle  = "3c2aec69-ffd9-4a34-888c-58adbbb99bb5"
	B30SavvyHunterOracle         = "602132c2-8ee8-41f8-bfac-cb17d32203f5"
	B30ExtractFromDarknessOracle = "e597d8a1-3bbc-4001-b642-f4421447970f"
	B30DocksideChefOracle        = "fed12a16-8920-403c-be63-0601a9d864b0"
	B30GrimGuardianOracle        = "c1f1babf-13d0-4fc4-b192-127d2d5db7f1"
	B30ChainOfSmogSkipOracle     = "ea14c26b-bf2f-48b4-b879-6e63069ded1f"
)

// b30Lives snapshots every seat's life total.
func B30Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b30TokensNamed counts the battlefield tokens `controller` controls
// with the given name.
func B30TokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && effects.IsToken(c) {
			n++
		}
	}
	return n
}

// b30Flyer seeds a non-catalog creature with printed flying.
func B30Flyer(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Keywords: []string{"flying"},
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b30DiscardOne discards one card from a player's hand through the
// effect path, so discard triggers see it. With one card in hand the
// random pick is that card.
func B30DiscardOne(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(player, 1); err != nil {
			t.Fatalf("DiscardRandomForEffect: %v", err)
		}
	})
}

// b30Reanimate puts a graveyard card onto the battlefield under
// `controller` through the effect path.
func B30Reanimate(t *testing.T, g *game.Game, card, controller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(card, game.ZoneBattlefield, controller); err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
	})
}

// b30SettleOrder answers a trigger-order prompt if one is open, then
// passes priority until the stack is empty.
func B30SettleOrder(t *testing.T, g *game.Game) {
	t.Helper()
	if TriggerOrderPrompt(g) != nil {
		AnswerTriggerOrderInOfferedOrder(t, g)
	}
	PassPriorityAroundTable(t, g)
}
