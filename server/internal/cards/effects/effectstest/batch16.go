package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch16_test.go — card-level coverage for the card-coverage
// roadmap's batch 16 (#309, `edhrec_rank` 1727–1829): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b16-prefixed.

const (
	B16DeathBaronOracle           = "99024aa8-5687-4d38-8a4b-feef42d6c1ff"
	B16FlameOfAnorOracle          = "ecd49d85-9c8c-4cc9-9a83-072ecd433677"
	B16SacredPeaksOracle          = "fb69bc57-f05a-41c2-9b7b-9a9761ef0cd3"
	B16VoltaicKeyOracle           = "09aeea91-b1dc-443f-a509-4758f052c0a7"
	B16RelicOfSauronOracle        = "b3a81bb1-cbd5-41a0-8dd4-faea06593f84"
	B16TombOfTheSpiritDragonOrcl  = "22f6391e-2634-440f-af1b-9581d1bff818"
	B16SunlitMarshOracle          = "a5e5a259-5fa7-4b01-93cb-a2b4aaf80927"
	B16DaxosOracle                = "9d3c7c96-056f-408e-a834-fa45a430d3d4"
	B16CircuitousRouteOracle      = "30afacd9-4680-4aac-8c22-584f9418822d"
	B16RiptideLaboratoryOracle    = "444d50dd-a44a-42db-bbf6-d0978e3bd6a3"
	B16UrabraskTheHiddenOracle    = "5b2ffb53-86b7-4665-a5c7-b85b035b6c81"
	B16TevalsJudgmentOracle       = "71fc2393-f55c-4b06-897c-fb7d4199b5f5"
	B16AureliaTheLawAboveOracle   = "2a800427-ff8c-4b3c-baee-85211b70656d"
	B16IzzetGuildgateOracle       = "bf75a3d1-f184-4b48-a913-21caee1db084"
	B16BartolomeDelPresidioOracle = "f47e4c56-0a0b-422d-bf3d-7a20ee289f15"
	B16VengefulBloodwitchOracle   = "4cdbc466-42fc-471f-beab-397caec18101"
	B16PsychicCorrosionOracle     = "328b42f1-d679-4f9c-80e3-38fe3b965d10"
	B16ThoughtScourOracle         = "83101ba8-a569-4827-8c53-9ca0dfcd59a7"
	B16GoblinMatronOracle         = "145737a7-c597-4dec-b752-207c2d0501e3"
	B16CollectorsVaultOracle      = "f460ff80-8175-43f7-9810-540cf3817085"
	B16DemonicCounselOracle       = "712e3479-722c-40a1-9b61-d5bdde93042b"
	B16LeadenMyrOracle            = "f62cabf0-df0d-4c4f-a93a-9340967d1775"
	B16ZacamaOracle               = "23270e4a-a222-4ffb-a946-d0e20d665187"
	B16DimirGuildgateOracle       = "52d14717-0cbc-4d7e-b546-54ea91580338"
	B16BraveTheSandsOracle        = "4e89bd75-f59d-4f08-be51-5660fbbba3c2"
	B16MurmuringMysticOracle      = "dcd4da46-5438-4454-8b1b-43ca51bda1f9"
	B16HorizonExplorerOracle      = "e8d20361-d9b7-4f9c-8ec5-3ac7c460dcb2"
	B16LyraDawnbringerOracle      = "592c91fc-6430-4c76-9460-65f047350f67"
	B16UnstoppablePlanOracle      = "3f6f4c98-ed7e-4fb1-8bfe-4210a39f77f2"
	B16DisallowOracle             = "88b51e15-6630-4e14-a6b8-db0aa12e34ef"
	B16WoodedRidgelineOracle      = "c2ca3e20-23ca-4d2a-88a1-5e98ff884abb"
	B16JuriOracle                 = "3364bad6-6d0a-4141-b410-86e3d9e1916e"
	B16TrueConvictionOracle       = "fc299c1c-50f3-492a-b6b8-a3664bb72ab7"
	B16BennieBracksOracle         = "17bd7ef7-8b4b-4a2d-a667-c751e10e2a47"

	B16SimicGuildgateOracle = "e8705df9-6439-4930-91b6-229f818559af"
)

// b16Creature seeds a non-catalog creature with a type line and
// colours, able to attack and tap.
func B16Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b16Activate activates ability idx on a permanent and settles it.
func B16Activate(t *testing.T, g *game.Game, controller, card uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, card, idx, params); err != nil {
		t.Fatalf("ActivateCatalogAbility %d: %v", idx, err)
	}
	SettleAnsweringProliferate(t, g)
}

// b16Tapped reads a battlefield card's tapped state.
func B16Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// b16CountNamed counts battlefield permanents with the given name.
func B16CountNamed(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// b16Tap taps a battlefield permanent directly.
func B16Tap(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
}

// b16TargetPlayer / b16TargetCard build target refs.
func B16TargetPlayer(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func B16TargetCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// b16PickPlayer answers a pick_target prompt with a player.
func B16PickPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}
