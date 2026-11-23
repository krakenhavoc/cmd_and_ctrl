package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch21_test.go — card-level coverage for the card-coverage
// roadmap's batch 21 (#383, `edhrec_rank` 2235–2334): the "no new
// machinery" group, the first slice of the second 2000. One test per
// observable behaviour, driven through a real cast, activation, land
// play, attack or step change. Helpers from the earlier batch test
// files are reused by name; new ones are b21-prefixed.

const (
	B21DemonsDiscipleOracle       = "d5a33091-a348-4b13-8dbd-79ab0ad99afe"
	B21LlanowarTribeOracle        = "cfd0f0e6-1bbf-4450-97d2-c54907abb7e4"
	B21IdyllicBeachfrontOracle    = "0aea7e5d-40d7-46c7-a8a6-479cb8061e49"
	B21SatyrEnchanterOracle       = "aa321138-b1a7-4b8e-a2ca-b9ce65704e92"
	B21EndTheFestivitiesOracle    = "09618da4-6e2d-4047-96d7-0576c0754e7d"
	B21ManglehornOracle           = "b67db32b-30a9-49d2-b4c2-90a9f80c36eb"
	B21RoyalAssassinOracle        = "9ed6f28f-a3db-48c5-9ab0-b90a7fba5f57"
	B21SeraphSanctuaryOracle      = "0b504dc6-61cc-4a72-907c-145fa4c72466"
	B21WrathfulRedDragonOracle    = "17c2a962-994a-49f1-8c4e-ebcc43c3c92a"
	B21GutShotOracle              = "9afb3b6e-4909-4efa-aa79-81c0229411c9"
	B21AgentOfTheIronThroneOracle = "325032d5-c452-4454-8976-82f86fee5ab8"
	B21SpitefulBanditryOracle     = "fde2653f-5270-4b8f-9642-0835dbb076c2"
	B21WillOfTheMarduOracle       = "d8df9813-0376-4d30-8cdc-30451fb67726"
	B21HermitDruidOracle          = "16f6438d-2a29-41cb-bf0c-4d02bd66112b"
	B21ShireTerraceOracle         = "619173f4-0403-49cd-9659-2fedd5028a90"
	B21MemniteOracle              = "7663ac7c-1de3-4250-b96a-fae9dbd66a27"
	B21TrophyMageOracle           = "08afc0d7-192f-4ab6-b6a0-c4265cf5e225"
	B21GeneralKreatOracle         = "da1c1c70-c2fa-4ba7-89fe-d9af3fb353b9"
	B21ElvishVisionaryOracle      = "c6a3a882-a127-4590-93d7-679ef4313efe"
	B21KokushoOracle              = "cf631c93-b7fb-4e4f-b405-3778c15b1117"
	B21MindsEyeOracle             = "63fd2a57-7a47-4e07-947c-f4e9da7ee538"
	B21LysAlanaHuntmasterOracle   = "3f3439e1-75ce-482d-881f-836492dca6e9"
	B21SpitefulVisionsOracle      = "922cf963-2b1b-43ad-819e-6e49133e6aae"
	B21LightningStrikeOracle      = "f34b9bc4-7bfe-47fd-ba23-4eeeb46026eb"
	B21DiregrafCaptainOracle      = "70de24d9-c585-4bf6-ac2c-c5b4b7aa298c"
	B21BrilliantRestorationOracle = "9584a8ae-2aba-42b8-8983-0467d6bd5698"
	B21VerduranEnchantressOracle  = "cd98a31b-cc7e-43f9-982e-109ad9850908"
	B21GethsGrimoireOracle        = "ef809e99-34a2-4471-8269-f56bf8037686"
	B21PrizedStatueOracle         = "681fc668-cb26-4ba4-a915-48ddfa2b9520"
	B21OblivionSowerOracle        = "d39b9f64-dc9b-413f-8d05-e21ef46d6756"
	B21SanctumSeekerOracle        = "afb71560-0fc9-4ea5-9d52-d93c17d72519"
	B21HazelsBrewmasterOracle     = "e8180024-1979-4677-9a0d-e08d4b7c825a"
	B21GeothermalBogOracle        = "e3b67368-1dd6-419b-a95d-7131b1dba23f"
	B21RestInPeaceOracle          = "087f9ad7-e74f-40e2-8102-1ed2925d0418"
)

// b21Push seeds a catalog permanent under `owner` with a type line
// and colours — b12Push with the colours a tribal or colour check
// might read.
func B21Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b21Commander seeds a creature flagged as a commander `owner` owns.
func B21Commander(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, IsCommander: true,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b21SettleStack is passPriorityAroundTable for a board where two
// differing triggers from one source arrive together: the CR 603.3b
// order prompt is answered in the offered order whenever it appears.
func B21SettleStack(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if StackFullyEmpty(g) {
			return
		}
		if TriggerOrderPrompt(g) != nil {
			AnswerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("stack did not empty after 64 priority passes")
}

// b21DeclineCommandZone answers the CR 903.9b "put it into the
// command zone instead?" optional_replacement prompt for a commander
// headed for hand or library with "no", so the card goes there. A
// graveyard or exile move is not asked first (answerCommanderReturn).
func B21DeclineCommandZone(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == owner {
			if err := g.ResolveOptionalReplacement(c.ID, owner, false); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			return
		}
	}
	t.Fatalf("no command-zone prompt for %s", owner)
}
