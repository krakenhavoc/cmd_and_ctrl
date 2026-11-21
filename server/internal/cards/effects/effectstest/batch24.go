package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch24_test.go — card-level coverage for the card-coverage
// roadmap's batch 24 (#386, `edhrec_rank` 2536–2637): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b24-prefixed.

const (
	B24MariOracle                = "5f1fdc23-9af0-41a1-aeba-7288f9642734"
	B24DebtToTheDeathlessOracle  = "6130f22d-7901-4f9f-b777-27bb0dacc063"
	B24NaturesWillOracle         = "9837287b-d821-4655-8f01-93fc6c7f0ecc"
	B24AgnaQelaOracle            = "22d0a848-2126-48f0-9050-38daaf93b1d0"
	B24IndulgentAristocratOracle = "c7054a2d-2e7b-4487-8fc3-f6a47a716fd3"
	B24MerchantScrollOracle      = "86cebe2a-95e7-4f22-99cc-e805aeaf347e"
	B24SpineOfIshSahOracle       = "02f062f5-8012-4440-ac12-49fc49822106"
	B24GoblinChieftainAlreadyOID = "368b4052-174e-4458-a6e6-eaf8093aa0fe"
	B24WitchOfTheMoorsOracle     = "7ad97dd9-1342-4ed4-bea9-1bb21d748e04"
	B24BugenhagenOracle          = "4228dd75-1fdd-48dc-9259-b841b2e46c64"
	B24KaervekOracle             = "c7b72d38-0aa0-4e17-9dd5-9276d7cb21ec"
	B24ScourForScrapOracle       = "da422b3e-1092-47ad-97d6-a7a3ef060a53"
	B24AjanisWelcomeOracle       = "4a782bf9-4051-4613-8852-33b0d85a0edd"
	B24MegrimOracle              = "633ad9e2-9f55-4a1c-9248-661ad4b0e1dc"
	B24SceneOfTheCrimeOracle     = "ba11a517-1dbd-4797-9f5e-46ce0f6c77c0"
	B24SurvivorsEncampmentOracle = "152e7e91-4eda-4e72-a9fb-bd5cb2e68239"
	B24SoulsFireOracle           = "62d7ed6e-c386-477e-b155-982c3790f842"
	B24AjanisPridemateOracle     = "95e94dea-5ac0-4d6f-adec-ca147aee861f"
	B24LichKnightsConquestOracle = "b0e533dd-baf2-482a-b9e1-872eb5e47439"
	B24VraskaJoinsUpOracle       = "c91b0dd5-4c63-49a5-95bf-c342b6ff2076"
	B24ChulaneOracle             = "ebf7ce9b-9e5e-4557-9e28-76556997f0ee"
	B24BloodSeekerOracle         = "41087db6-34c4-4e2b-9f54-5e4488ca9c0b"
	B24BurglarRatOracle          = "2f807301-37df-4724-871a-08e3512b07b3"
	B24FinalPartingOracle        = "a5852994-d816-4e62-8a03-254223714544"
	B24WillOfTheAbzanOracle      = "1ae29791-aa7c-4050-bf72-dd0f739b11b8"
	B24MoltenTributaryOracle     = "58c592ed-20fc-481b-909b-2315567e5f20"
	// #1210: both were declared skips for "no activation gate", the
	// gate landed, so they move up into `want`.
	B24LinvalaOracle               = "88dadc31-dfac-41b3-bf2a-65fa89e3c16d"
	B24CursedTotemOracle           = "6225a704-430a-4f56-ad87-0e8d87f285f5"
	B24ThroneOfTheGodPharaohOracle = "ea750169-1f6f-40c2-96e9-55719e103a63"
	B24WrathfulRaptorsOracle       = "1ba92b64-d821-4c24-aed6-fc90d0c63c10"
	B24EfficientConstructionOracle = "5af48f87-7b94-44de-90e3-91f10ced00d3"
	B24TectonicHazardOracle        = "9097243a-39fa-4e18-8316-c0e57699c783"
	B24DwynenOracle                = "30d0d75f-e94c-460b-b957-9f1d655c0f65"
	B24YshtolaNightsBlessedOracle  = "3268251a-8292-44f9-9267-c961b182f739"
)

// b24Lives snapshots every seat's life total.
func B24Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b24CastNoncreature puts a noncreature card with a real mana cost in
// the active seat's hand and casts it from a main phase.
func B24CastNoncreature(t *testing.T, g *game.Game, name, typeLine, manaCost string) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := HandCardFull(me, name, typeLine, manaCost, "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
