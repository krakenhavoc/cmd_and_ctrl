package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch28_test.go — card-level coverage for the card-coverage
// roadmap's batch 28 (#390, `edhrec_rank` 2940–3040): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, block, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b28-prefixed.

const (
	B28AbradedBluffsOracle         = "ca7d093c-0533-493f-9ad3-8af30118fbfc"
	B28DoppelgangOracle            = "d04c6375-25dd-4882-b8b3-a3b0d3081f32"
	B28ReviveTheShireOracle        = "67afcfef-3f0c-4823-b690-01732f28a005"
	B28UncivilUnrestOracle         = "bd655e8b-f192-4635-9e23-357b6f89ef8f"
	B28NykthosParagonOracle        = "367ac7e2-5056-48f3-a296-87d6cb1f7b54"
	B28BenevolentHydraOracle       = "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0"
	B28PhyrexianWalkerOracle       = "7af75024-6c9b-4844-aeb6-81de25464822"
	B28HinterlandSanctifierOracle  = "d812fc6d-b96d-4986-b171-9f3feee603dc"
	B28UlvenwaldTrackerOracle      = "63a37c08-7e38-4fe2-9a71-cc4604c4f831"
	B28CauldronFamiliarOracle      = "a7dc2e62-1c50-4ed7-b71f-2d782a447a5e"
	B28DragonstormGlobeOracle      = "f6de5bd7-7704-4a0c-a27a-9e565e49f5e9"
	B28RepercussionOracle          = "f6069a4f-744e-43e8-9f8b-7c13b8b0187f"
	B28FatedFirepowerOracle        = "13daa21c-278d-45bd-9a6e-a77d6a558453"
	B28RegalCaracalOracle          = "aa571676-b390-4b8a-baea-4c4cd60c01f6"
	B28TiamatOracle                = "f00e4df1-13fb-4514-8abb-92954068689a"
	B28ArcanisOracle               = "8a7183cc-161c-444d-a889-a17519c8061b"
	B28MizziumMortarsOracle        = "48ddba1e-2ad7-463f-9307-d2379a800e51"
	B28BeseechTheQueenAlreadyOID   = "cf94cafc-527e-4b27-8a28-7807435aaccf"
	B28KykarOracle                 = "71a80491-eea2-4661-9446-26a87efbf4a8"
	B28RavenousSquirrelOracle      = "8fb48d65-406b-4231-83ea-9cf3bb8dff76"
	B28KarametrasAcolyteOracle     = "e8cf80a1-6908-4adb-88e7-2015672d4905"
	B28ThunderbreakRegentOracle    = "bb04f927-0348-4a14-9e74-381f18083f6a"
	B28ChivalricAllianceOracle     = "9b8d9236-0452-4509-aca4-a53a399fff85"
	B28MarchOfTheMultitudesOracle  = "0c26ab0d-80f6-4e5b-9d0e-af17c1519583"
	B28BrimazOracle                = "49471b65-be9d-4ded-b2e5-dfc81b306b54"
	B28StromkirkCaptainOracle      = "599c3fd6-3309-4b5d-adec-9c4062848ad5"
	B28LifebloodHydraOracle        = "b14d05c0-fe10-4079-a90e-0aea1a8fd375"
	B28SpawnbedProtectorOracle     = "256a7f54-0e8e-4d22-a0f7-ff2830e8884e"
	B28HoldoutSettlementFullOracle = "e6b77545-de5c-4f4a-b7ea-83498fb33ba8"
	B28FertilidOracle              = "21f1c6d7-8289-44b2-b88f-c09e202be200"
)

// b28Lives snapshots every seat's life total.
func B28Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b28TokensNamed counts the battlefield tokens `controller` controls
// with the given name.
func B28TokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && effects.IsToken(c) {
			n++
		}
	}
	return n
}

// b28AddCounter puts n +1/+1 counters on a permanent through the
// effect path, so replacements see it.
func B28AddCounter(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, "+1/+1", n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

// b28TapForMana activates a permanent's first mana ability and
// answers every colour pick it opened with `color`.
func B28TapForMana(t *testing.T, g *game.Game, controller, card uuid.UUID, color string) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if color != "" {
		B10ResolveAllManaPicks(t, g, controller, color)
	}
}
