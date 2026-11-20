package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch15_test.go — card-level coverage for the card-coverage
// roadmap's batch 15 (#308, `edhrec_rank` 1626–1726): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b15-prefixed.

const (
	B15WhirlwindOfThoughtOracle    = "6467cbb7-1e4e-482d-a20f-6cb9fc0f1ad1"
	B15StitchTogetherOracle        = "bc3d5911-3580-4132-9daf-2826495b5739"
	B15SoulOfTheHarvestOracle      = "6b9194dd-2296-4879-ac5e-f89b18431df6"
	B15ObscuraStorefrontOracle     = "dc31a6f8-6228-4a25-b937-5d8d78514333"
	B15WindsOfRathOracle           = "a6fd90dc-0ec3-4dce-a77a-4d04f5e254bf"
	B15PoisonTipArcherOracle       = "d0e810bb-5f38-4045-a718-30d423c05659"
	B15BloodMoneyOracle            = "75f5d372-4ff9-430c-8302-72472439e0d2"
	B15SilentClearingOracle        = "fd45063f-c83c-431c-9104-f139c497ec0d"
	B15InspiringOverseerOracle     = "d646e42b-5635-4798-b633-29c093b66a55"
	B15BotanicalSanctumOracle      = "88f8f683-738e-48f3-afff-c8f73f1033a2"
	B15RunawaySteamKinOracle       = "eb414a8f-566f-4635-80a1-77b2e18ac8dc"
	B15MoonsilverKeyOracle         = "373395b2-f04b-46c9-b9f4-4f6815b71009"
	B15TheGafferOracle             = "a177295e-3b58-4e46-a1cb-fc003a7a0848"
	B15RunAwayTogetherOracle       = "290faa28-450e-4797-9a8f-642d8af3f82a"
	B15UtterEndOracle              = "cba94732-1f1f-4ffd-9aba-45db990043fa"
	B15SmugglersShareOracle        = "17b29350-4f37-4552-8192-4856b15345f9"
	B15SunscorchRegentOracle       = "c8ff464f-0059-4055-a2f2-556fe6db8fbf"
	B15HorizonCanopyOracle         = "262a5d83-506c-4781-9bc9-1a2b5d83955c"
	B15HornetQueenOracle           = "3b1f8108-6911-49e9-8f78-f950bb58cb6c"
	B15OversoldCemeteryOracle      = "ed4cd3a1-688b-4c05-948d-39d3336e00c0"
	B15MirrorBoxOracle             = "3bed1944-58dc-4679-9aee-7be4d94fb55c"
	B15KherKeepOracle              = "79638767-fbc7-451a-b29f-d93f2ac6f102"
	B15GalaGreetersOracle          = "cce081eb-8820-415a-a7b2-3c5b9d4a2601"
	B15FateUnravelerOracle         = "7d66f67d-3148-4636-ac5a-2cf8a51d5e50"
	B15ThermoAlchemistOracle       = "228fbae1-423e-461d-b8c3-55786938a3cb"
	B15SimulacrumSynthesizerOracle = "eb7a1f21-a66d-415b-8520-710b44890bb6"
	B15TheLocustGodOracle          = "e025a714-02da-4b0c-8021-cf3e8dc9b19e"

	B15SolRingOracle = "6ad8011d-3471-4369-9d68-b264cc027487"
)

// b15Destroy destroys a battlefield permanent through the effect API
// and settles whatever it triggered.
func B15Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	PassPriorityAroundTable(t, g)
}

// b15Aura attaches a fixture Aura to a battlefield creature.
func B15Aura(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment — Aura",
		Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

// modePickChoiceFor is the open mode_pick prompt for a chooser, or
// nil.
func ModePickChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceModePick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
