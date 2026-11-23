package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch09_test.go — card-level coverage for the card-coverage
// roadmap's batch 09 (#302, `edhrec_rank` 1005–1110): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b09-prefixed.

const (
	B09DismalBackwaterOracle    = "865a2194-fca0-446e-aae3-ca475cd66e00"
	B09WindScarredCragOracle    = "b0af0c54-2a59-4075-8543-d41ff20c4c87"
	B09SwiftwaterCliffsOracle   = "2f4ad084-2062-44c0-9975-15f100204531"
	B09BlossomingSandsOracle    = "45429b2c-be3b-4b2e-9bab-a059ccbda8cd"
	B09ThornwoodFallsOracle     = "ec96cde2-f1e6-495c-94e2-3e8ae79e556c"
	B09TranquilCoveOracle       = "5d641bf6-0f93-4189-8dc1-ec7ea446dade"
	B09CastleVantressOracle     = "cdf41cf4-4e77-453d-be5b-0abbbd358934"
	B09TalismanOfUnityOracle    = "e5fcc5d7-6a60-4a5b-9d02-6c30041a95b9"
	B09RewindOracle             = "bb27bfdf-fe8d-45bd-ad62-8118dce06eda"
	B09UnwindOracle             = "e38b8ecb-e7ae-474d-b6a4-29cc1aa8ccd9"
	B09ArastaOracle             = "695eea46-1535-48c5-bbb6-0b8379e77bfc"
	B09SetessanChampionOracle   = "38513660-260f-4663-b199-6473fa004a3b"
	B09HourOfReckoningOracle    = "3bc13640-03f8-4b19-b0a4-7e7cb5271c0c"
	B09StaffOfDominationOracle  = "d7888719-647d-4022-a211-822fa09f0791"
	B09SteelshapersGiftOracle   = "d9abda7e-6ca2-42ea-ab24-c542e57014f1"
	B09SpellseekerOracle        = "47a785ed-8095-4685-8daa-02c4e2b0ffcd"
	B09StrixSerenadeOracle      = "7ca3aa03-e62d-464a-9111-e754abe17f76"
	B09SyphonMindOracle         = "abc37d6c-6300-47b5-a679-9db5b83eb54f"
	B09RoilingRegrowthOracle    = "65d63518-3261-40b5-87b8-19c152b29ee3"
	B09BlackcleaveCliffsOracle  = "5ad94412-6f79-4c5d-bbd4-4ef5779a7b6d"
	B09JhoiraOracle             = "c803b788-4213-4fab-b841-7e5bbf66088e"
	B09WallOfOmensOracle        = "5f601f48-d24b-4883-9fde-b3f620e7c9ea"
	B09SharedRootsOracle        = "9e0fd3bf-f47a-4f06-8ff1-73f6bf5d1e03"
	B09DisenchantOracle         = "a7e97fa9-4b72-4548-b854-5be5f18a6f1a"
	B09ArchonOfCrueltyOracle    = "aa1a6646-c1e6-4bff-9092-43ee3e137914"
	B09StoneforgeMysticOracle   = "358789f9-7d87-411d-919e-d597da665cbd"
	B09UpTheBeanstalkOracle     = "050f5733-7c0b-4991-9a6c-7ea12ccf0ca9"
	B09FumigateAlreadyOnMainOID = "b17ea905-0696-4e58-b564-557e87236e27"
)

// b09OpponentCastsCard is batch01OpponentCasts for an arbitrary card:
// the non-active player casts it at instant speed. A creature gets
// flash on the fixture so the sorcery gate lets it through.
func B09OpponentCastsCard(t *testing.T, g *game.Game, opp *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = opp.ID, opp.ID
	opp.Hand.PushTop(c)
	if err := g.CastSpell(opp.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// b09AddColorless floats n colorless mana in a player's pool.
func B09AddColorless(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}
