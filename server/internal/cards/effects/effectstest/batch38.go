package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch38_test.go — card-level coverage for the card-coverage roadmap's
// batch 38 (#401, `edhrec_rank` 3953–4052). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives directly.

const (
	B38TurbulentMoorOracle       = "2eb4da30-2600-4a7f-8e6c-6a090faa9a8d"
	B38CreosoteHeathOracle       = "c116b787-5f7e-47ef-a694-58709770dd32"
	B38HedronCrawlerOracle       = "0b9a4e06-b21d-4cbe-906f-9dbd08dbe5d3"
	B38AlloyMyrOracle            = "efb0394c-2a45-4dd8-bca3-08704056fa31"
	B38KingOfThePrideOracle      = "96d0e4dd-6cc2-4349-ac44-785b50f8dd90"
	B38GaleriderSliverOracle     = "1ad78038-44f9-4599-84db-fc1a87d9ed45"
	B38ArmageddonOracle          = "c9ed8b01-959a-47d6-891e-0abbdccf6e4f"
	B38InnocentBloodOracle       = "6791ec3c-c397-4087-8c8c-84d3797df415"
	B38AnnoyedAltisaurOracle     = "a8135179-51ef-454e-98ff-69137440339f"
	B38VirusBeetleOracle         = "e1e489f6-37e5-469e-bca3-0e9d3fc3fa95"
	B38ObeliskOfUrdOracle        = "e2ef3a24-9e78-47fc-9192-049aa0ddb7a0"
	B38FlameRiftOracle           = "28479aa8-d1cd-421a-8bbb-0594bb4dd410"
	B38PestilenceOracle          = "dafe63ef-f3d6-45e7-877a-573da92ba85e"
	B38FirespitterWhelpOracle    = "1cad92ce-55c8-4b78-8d8c-56645ef8e6ee"
	B38SpiderManifestationOracle = "8281f2b9-e81b-48da-812a-8713b2adf8ab"
	B38DesolationTwinOracle      = "c0cb1f37-1679-42c7-a794-81e088157eeb"
	B38ViridianRevelOracle       = "c24b0c7d-1a4b-4eca-8d83-9e81ed3b3b27"
	B38HydraOmnivoreOracle       = "8f504855-f3df-4284-a189-e799bcddf620"
	B38HareApparentOracle        = "3c1619bd-db5e-4df6-a196-0a9d62374f6d"
	B38ColossificationOracle     = "05c5ee65-651b-4bfb-b57f-305026507135"
	B38HerosHeirloomOracle       = "f707dfb5-dfa9-471b-ac29-5435c30069cc"
	B38DoomedNecromancerOracle   = "155422a0-a0cd-4399-8ed9-fa68ac2c80a6"
	B38PhyrexianDelverOracle     = "a13cbac0-4c76-4970-b61e-5f4e020ee95c"
	B38QuicksilverAmuletOracle   = "e1096a98-a631-43f5-97d8-325001d608b3"
	B38ArtfulDodgeOracle         = "c174dcbb-03a0-439c-b3d8-ed61bd46dc67"
	B38ExhumeOracle              = "fbe61f74-1b3c-4e12-8758-7029872c9ff1"
	B38SharedSummonsOracle       = "c2c9dbd0-2062-4ee1-b32e-8eeacd95589c"
	B38RelmsSketchingOracle      = "dd6601c6-810d-4883-89d2-5d31419fb1cc"
	B38HeartwoodStorytellerOracl = "82e6da87-f8a4-4897-bf0f-c0f2cd06b8b1"
	B38GrimServantOracle         = "1caa7fa2-a881-4a65-a87e-12a974520c83"
	B38MerrowReejereyOracle      = "6b7e6ae4-2ee1-44bf-ac93-79fc87494515"
	B38CanoptekScarabOracle      = "5b7f6cc3-8d4e-42e5-a908-7ebe3bccaf1e"
	B38TitaniasCommandOracle     = "7aae0a3d-8882-4485-a126-f06ea6593dcb"
	B38SludgeMonsterOracle       = "2802fc92-49a0-43c8-bc10-24882eeee3f1"
	B38OpenTheVaultsOracle       = "1e9c473e-bd65-4e1f-b2ab-cac58dc581c9"
	B38SpellStutterOracle        = "f9bf996e-ae14-40a3-a8d8-f5725ddaa270"
	B38WilsonOracle              = "d2766fd7-5cf9-4037-9f34-9ae3982c613a"
	B38VinesOfVastwoodOracle     = "8998d211-5ca6-41d2-bc64-f51f70bd41e5"
	B38DruidsDeliveranceOracle   = "fde7645a-5f02-4d5f-b38c-8390f325899e"
	B38ManaSculptOracle          = "35e2f82e-7ca3-4a92-9134-b7999eef5337"
)

// --- shared fixtures -----------------------------------------------

// b38PT reads a battlefield card's post-layer power and toughness.
func B38PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	p, tough, found := 0, 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				eff := c.Effective()
				p, tough, found = eff.Power, eff.Toughness, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return p, tough
}

// b38CastCreature casts a creature with a real body from the active
// seat's hand and settles the spell and whatever it triggered.
func B38CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// b38Graveyard seeds a card into a player's graveyard with a real
// mana cost, so a mana-value rider has something to read.
func B38Graveyard(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	p := g.PlayerByID(owner)
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
	return id
}
