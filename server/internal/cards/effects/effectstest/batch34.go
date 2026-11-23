package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch34_test.go — card-level coverage for the card-coverage
// roadmap's batch 34 (#397, `edhrec_rank` 3550–3649): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	B34InvigoratingSurgeOracle    = "5daac63a-1534-4194-8cb7-506508e364f4"
	B34ReveillarkOracle           = "1be13ede-98f8-497e-800c-03e5802932b3"
	B34YavimayaElderOracle        = "7fd4c452-07f2-492c-9c78-d1c6362d9eec"
	B34VitalizeOracle             = "cd3226c6-bd47-4246-8ab3-01316c5d1809"
	B34FiendishDuoOracle          = "ab0dfae5-b9d4-417b-8a0d-2525ae3a73b9"
	B34ArmorcraftJudgeOracle      = "d7f49243-a96e-499f-b2d7-8e9842432420"
	B34OmnathLocusOfTheRoilOracle = "ffd34151-457a-4a93-82db-24b18319a06b"
	B34BeetlebackChiefOracle      = "f5c5f64c-6911-430c-a825-b32b96d39c7d"
	B34CoriMountainMonasteryOrcl  = "35c60b66-8c85-432e-90fe-99c19d21ed15"
	B34BristlingBackwoodsOracle   = "9cbc9f83-8979-42a5-a466-a8d89c8e6de8"
	B34SporemoundOracle           = "1be56a3d-a6c0-4b65-ae71-3d90ceefc6c0"
	B34GuildArtisanOracle         = "aceac269-5100-428a-a4a3-bac57031e30b"
	B34LordOfTheUndeadOracle      = "7714af0e-41d9-4609-967f-27233b46055f"
	B34AuroralProcessionOracle    = "511d3de9-076b-4811-bddf-00623f98993f"
	B34KioraTheRisingTideOracle   = "5d748bce-dff8-46fa-a3d1-633863b7bbff"
	B34RegalForceOracle           = "e264ffe3-0252-49d6-b990-dbb3654325a5"
	B34TurbulentFenOracle         = "114dd40d-5ad8-4913-a08f-572b9521eb5b"
	B34OnWingsOfGoldOracle        = "3b6f608d-f27d-44cd-b0a1-0e1e1aa1c98a"
	B34DragonbornChampionOracle   = "9e51ccf9-d55d-4231-8ff5-ef3ca1485c8d"
	B34ArbaazMirOracle            = "a383ef16-1af8-4b3a-956c-c10a93768617"
	B34MaelstromPulseOracle       = "95ce305f-34bc-4d6d-b7ba-ffd4b2a25336"
	B34SeizanOracle               = "04d0d20f-720e-4cb6-a3ad-ea2b57bb7efa"
	B34TormodOracle               = "6dc0150d-7145-41e2-bd3d-2564d9d32301"
	B34MapTheFrontierOracle       = "11a46cb6-ab01-4630-9541-782db2ef3b91"
	B34RapaciousDragonOracle      = "0944ec2e-1dd9-459f-8f1d-667242cf52fe"
	B34CrosswayTroublemakersOrcl  = "1a362e4d-6c02-4b67-ab63-c6622e505195"
	B34DeathsproutOracle          = "793b0c73-601c-41dc-b49a-45fee4970d52"
	B34ProsperityOracle           = "c586312d-d04a-4bfb-bbb2-b41186ca178e"
	B34HobgoblinBanditLordOracle  = "43fea418-db6f-4953-89d1-6873b2ca41a8"
	B34ElanorGardnerOracle        = "ad4c39d6-a8b3-4dd9-816d-09761fc9651d"
	B34VillageBellRingerOracle    = "d7b50175-b473-452a-bb30-19a71c42e649"
	B34PollutedBondsOracle        = "13b33a41-c22b-4e7a-8323-abadf7f80961"
)

// b34Changeling seeds a creature with the changeling keyword — every
// creature type at once.
func B34Changeling(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling},
		Power:    1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// b34LandEntersByEffect puts a land onto the battlefield under `p`
// through the graveyard-return path — an effect's entry, with a real
// EventETB, rather than a land play.
func B34LandEntersByEffect(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	id := B17GraveyardCard(p, name, typeLine, "")
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield); err != nil {
			t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
		}
	})
	return id
}

// b34PlayLandAs seeds a land in `seat`'s hand and plays it from that
// seat's precombat main phase.
func B34PlayLandAs(t *testing.T, g *game.Game, seat int, name, typeLine string) uuid.UUID {
	t.Helper()
	AdvanceToMainOf(t, g, seat)
	p := g.Seats[seat]
	id := HandCard(p, name, typeLine)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}
