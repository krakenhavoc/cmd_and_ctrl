package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vis_5631_cards_test.go — a batch of Visions cards (#5631). Every test
// plays the real catalog entry through the engine.

const (
	VisArchangelOracle             = "9971697b-2acc-4bc2-a44e-074d03a51df7"
	VisFreewindFalconOracle        = "8cdda36f-5e82-4c2f-8eaa-75b9bbc93a02"
	VisLongbowArcherOracle         = "51531db1-ba59-4ccc-9a91-54de2887fc8b"
	VisPantherWarriorsOracle       = "ef7788af-8edc-46df-a5b6-895c734ea423"
	VisPythonOracle                = "d8a0c3ff-7042-4b52-a216-15e170c8094f"
	VisScalebanesEliteOracle       = "c04dc6fe-06fc-4e5b-87da-6b8b262a214f"
	VisTempestDrakeOracle          = "ff0086e0-706f-4474-9b5e-a1591235bf9b"
	VisWarthogOracle               = "ae1f22a5-89cf-440c-b90d-fed194b2dee5"
	VisSuqAtaLancerOracle          = "f348b326-8a73-4df9-8faf-e271a36ce373"
	VisFallenAskariOracle          = "4956008e-d912-48ba-9b82-176258281060"
	VisHulkingCyclopsOracle        = "04936918-7ef3-42fa-b0cd-87d18cb69d2c"
	VisRiverBoaOracle              = "86187aed-77ce-4b2d-aab0-0a8f807a9451"
	VisWakeOfVulturesOracle        = "2948b1d5-c12c-45c4-a3e8-aa470fcee1af"
	VisDarajaGriffinOracle         = "daa0d720-cac1-487a-93e2-33e18f91f19b"
	VisViashivanDragonOracle       = "86fb1fc1-7784-4c94-a80c-21ad87e50fc4"
	VisKeeperOfKookusOracle        = "97417270-e525-4261-b7d9-0dde3360e666"
	VisJamuraanLionOracle          = "b558541f-ae57-41f7-a5bd-e23fa35e68b4"
	VisInfantryVeteranOracle       = "42798e2b-9a5c-4f93-8f77-b7bb7a916d07"
	VisSisaysRingOracle            = "8f5822ae-651f-410e-9316-5522eeb52d72"
	VisRainbowEfreetOracle         = "7c158405-d3ef-4c10-a504-7b061ef71a74"
	VisTeferisHonorGuardOracle     = "8a3a8a02-b7da-4648-a772-c9efb7c332da"
	VisCreepingMoldOracle          = "59180e94-ccdf-4d9f-9a4a-fe55497d0d63"
	VisElvenCacheOracle            = "8a922366-ee2a-4ddb-a904-683dcf4f191a"
	VisInspirationOracle           = "8f32ceb2-92c2-4dde-bf73-40bb79c3fcef"
	VisUndoOracle                  = "b446c3bd-3c7d-4d59-ba59-3dd80818465e"
	VisWarriorsHonorOracle         = "c3c9da49-ef4f-48c7-99c8-0f226b485285"
	VisMiraculousRecoveryOracle    = "cea73a4b-e10a-4517-a92b-24a08321009f"
	VisRetributionOfTheMeekOracle  = "61f2083b-3018-43a4-8852-8e6d2a97d5d6"
	VisTremorOracle                = "4281a153-665b-44cf-bcff-9cf88c14c2b5"
	VisManOWarOracle               = "67a3541c-8408-40c8-b44f-90035b860f57"
	VisUktabiOrangutanOracle       = "6cc07a7e-e70c-45c5-83c5-a2704e57dbc5"
	VisNekrataalOracle             = "d40adf5d-5edf-48a0-98f3-fcdbf1b5f25c"
	VisCoralAtollOracle            = "3f347ebf-e0d2-4ae0-ad84-df7a460404e0"
	VisDormantVolcanoOracle        = "38ba1956-5505-4a7a-b6af-e75715b1401f"
	VisEvergladesOracle            = "ef3b8b0c-cea7-4bae-934c-9c65fd64245d"
	VisJungleBasinOracle           = "f922f90a-b1a2-4630-9266-40726ca89f74"
	VisKarooOracle                 = "d4e875d9-2245-470d-aa2f-1dfe66ce2d15"
	VisHumanKnightTypeLine         = "Creature — Human Knight"
	visShieldedBoaTypeLine         = "Creature — Snake"
	VisLightningBoltForKeeperLabel = "Lightning Bolt"
)

// visShield gives a River Boa pushed at the given size a regeneration
// shield from its own ability.
func VisShield(t *testing.T, g *game.Game, owner *game.Player, power, toughness int) uuid.UUID {
	t.Helper()
	boa := B12Push(g, owner.ID, "River Boa", visShieldedBoaTypeLine, VisRiverBoaOracle, power, toughness)
	HmlSetColors(g, boa, "G")
	ApaMana(owner, "G")
	if err := g.ActivateCatalogAbility(owner.ID, boa, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("River Boa's regeneration: %v", err)
	}
	PassPriorityAroundTable(t, g)
	return boa
}

// visBattlefieldNamed is the first permanent named `name` that `owner`
// controls, or false.
func VisBattlefieldNamed(g *game.Game, owner uuid.UUID, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}
