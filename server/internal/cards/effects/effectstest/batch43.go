package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch43_test.go — card-level coverage for the card-coverage roadmap's
// batch 43 (#450, `edhrec_rank` 4456–4555). One test per observable
// behaviour, driven through a real cast, activation, attack or land
// play rather than by calling primitives directly.

const (
	B43GigantosaurusOracle      = "e666bae7-dd51-4921-8b89-7e8d423caba0"
	B43IndomitableAncientsOracl = "2764741c-1f3e-459a-a487-54cbac2c84a6"
	B43RaiseTheAlarmOracle      = "5b2364d7-a811-4595-a1b4-224c70555ffa"
	B43DreadboreOracle          = "d685799f-cc1a-40d6-9df6-d05b8b1f5b13"
	B43NyxFleeceRamOracle       = "5a20113c-7ea8-4edf-af9f-148ccd326a88"
	B43BaronOracle              = "cc710da0-5a2e-4bc4-8fdd-d90e7bc1f224"
	B43ThornglintBridgeOracle   = "99720c65-be96-4220-8ed4-720660bf6928"
	B43UrzasFactoryOracle       = "5a620d20-f14e-43d0-8e57-c2a197e2ec51"
	B43NayaPanoramaOracle       = "71e28800-c42c-48c0-95e5-0296be54a4e8"
	B43ShamanOfThePackOracle    = "39c7d2cc-4fd1-4f1a-be1a-a6ae56d2356d"
	B43SeafloorOracleOracle     = "f9f719cb-7778-4109-bbd3-4504ee1b5f49"
	B43NemesisOfReasonOracle    = "b1d70845-869b-45f7-b4bf-1549bac51d58"
	B43PeekOracle               = "873abbf4-c9b5-4b8f-8cd0-613cf3b9b1d5"
	B43CaptainSisayOracle       = "2e7a9ea9-f76c-4c12-950f-c613fa16cfa8"
	B43PlanarBridgeOracle       = "853c6cc0-ed2b-4d65-add1-2449ade8cf68"
	B43NightshadeHarvesterOracl = "875c3db1-2752-48e7-9664-bbb12120c032"
	B43DarienOracle             = "d05336cb-6157-47e7-942f-43becd67a5bf"
	B43ElectrickeryOracle       = "99fd4b51-1698-4168-9366-a6ea2df22361"
	B43AngelicFieldMarshalOracl = "502575f4-7c56-44bc-a77f-ae28d66c8e1f"
	B43BoonOfTheSpiritRealmOrac = "7731217e-cdbb-4c38-a9c1-083c96f51931"
	B43NoxiousGhoulOracle       = "602e421e-afab-4488-8a52-af66b4fdce4b"
	B43DistinguishedConjurerOra = "6df57d67-2fd9-4e7a-b67b-f361fc30e496"
	B43CeruleanWispsOracle      = "a6605c50-558e-417c-8c75-6c45b06d6e13"
	B43TowerWinderOracle        = "c1aecc8a-db3a-4158-b1f1-bfe06d782482"
	B43MooglesValorOracle       = "9627ed9a-5ade-412d-bf8d-f0705b1c9405"
	B43EowynOracle              = "98a2a642-a30a-43e7-b47f-4776c22d1cb0"
	B43DeeprootPilgrimageOracle = "188656d8-4ea8-4e58-88da-b010a16eb6f2"
	B43FinneasOracle            = "69924636-138e-4baa-a378-0fd7df5b847d"
	B43CalderaPyremawOracle     = "3eb5ae68-7f76-4874-80a4-80058b10fc7a"
	B43ZukosExileOracle         = "5cf18398-9552-415e-a61f-2c149e35ec3d"
	B43DragonlordAtarkaOracle   = "daff70a4-a990-47c9-b6ef-2c379272c8a3"
	B43MaelstromColossusOracle  = "06fc3eb1-7a48-42a0-8a5c-d61bd5aacfff"
	B43ZephyrBootsOracle        = "02f3b9c3-f611-4f05-ab3e-b296916efdad"
	B43ShieldOfTheOversoulOracl = "9b598025-80a3-4144-be9a-863165161594"
	B43DesertOracle             = "195107ad-879d-4b02-a44a-a3ba70fedf88"
	B43IfnirDeadlandsOracle     = "af698bd5-5f56-4d2a-9f02-8c3e781210cd"
	B43CryptcallerChariotOracle = "2bdd0bcb-6cfb-48e7-972e-355ff46621a4"
	B43ArchetypeOfCourageOracle = "79b48704-480d-4905-b87a-40b127894670"
	B43VizkopaGuildmageOracle   = "f19e7c5c-67fa-4ae4-89b8-afa0e08a6c48"
	B43RestartSequenceOracle    = "c415bfdd-3d42-4f94-9f93-7e310becbc8b"
	B43FracturedIdentityOracle  = "1515b0c2-1b55-4cd4-ad81-fb6b1f3e8188"
	B43FloodOfTearsOracle       = "3d92f2e6-27df-4329-b552-cf905f7616ba"
	B43AgonasaurRexOracle       = "1ad5766b-9ae3-437b-bd91-c4c988a8b095"
	B43InterplanarBeaconOracle  = "073169f2-da3a-4a93-8c01-b3fd8558d225"
	B43SkemfarShadowsageOracle  = "496ecb73-f508-4ac7-b101-23e03404580f"
	B43GimlisRecklessMightOracl = "824abca7-b5e2-4ddb-ac7d-b7d04758a3c9"
	B43LostJitteOracle          = "a9d72f78-2ab5-4e2e-ab7b-ef875e0a0609"
)

// --- fixtures ------------------------------------------------------

// b43Creature seeds a creature with a real battlefield timestamp, so
// the layer cache and the "entered this turn" tally both see it.
func B43Creature(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Colors: colors,
		Owner: controller, Controller: controller,
	})
}

// b43Catalog seeds a catalog permanent with a real timestamp and a
// real power / toughness — pushCatalogPermanent stamps 1/1 and no
// timestamp, which is wrong for anything the layer engine reads.
func B43Catalog(g *game.Game, controller uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: controller, Controller: controller,
	})
}

// b43TokensNamed counts battlefield tokens with the given name under
// one controller.
func B43TokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && c.IsToken() {
			n++
		}
	}
	return n
}

// latestPutFromHandChoice finds an open "put a permanent card from
// your hand onto the battlefield" prompt for the chooser.
func LatestPutFromHandChoice(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Chooser == chooser && c.Kind == game.PendingChoiceChooseCards {
			return c
		}
	}
	return nil
}

// b43BattlefieldByName finds the first battlefield permanent with the
// given name — the only handle on a permanent that has been blinked,
// since the return is a new instance.
func B43BattlefieldByName(g *game.Game, name string) (uuid.UUID, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID, true
		}
	}
	return uuid.Nil, false
}
