package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch40_test.go — card-level coverage for the card-coverage roadmap's
// batch 40 (#403, `edhrec_rank` 4154–4253). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives.

const (
	B40TurbulentWildernessOracle = "bd8adca6-4f16-45f8-994a-fe55bd573bd0"
	B40TurbulentSteppeOracle     = "db444f9d-4dde-4308-b0f2-7acfe6de871a"
	B40IceTunnelOracle           = "40c5d6fe-854a-436d-9f80-13eb5f1f8f68"
	B40GrandColiseumOracle       = "1cea9b82-d2e9-4758-8ec8-729fcf4bb7d7"
	B40PollutedMireOracle        = "9809d975-7ef8-4946-9041-607c4e954b13"
	B40RemoteIsleOracle          = "24aebda0-315f-4d2f-8bd9-00bbaf5bd76a"
	B40VisionSkeinsOracle        = "6d42dff5-97ec-4768-b112-84a3584d0b87"
	B40RevitalizeOracle          = "b1385b03-cb4b-4812-857f-7421f1df39af"
	B40MassCalcifyOracle         = "3ab3996f-aa3f-4041-8634-8e197d51f108"
	B40ShootTheSheriffOracle     = "2038138e-2e52-41e4-90fc-443fa054c5c0"
	B40LeapOracle                = "e89a3ce0-6b38-4326-a7af-8575af371baa"
	B40ManakinOracle             = "d2343af0-468b-42bc-8a0c-347c10f7e2f3"
	B40PatronOfTheArtsOracle     = "47f8df80-abc9-4d46-8078-8fbc23430259"
	B40DrogskolReaverOracle      = "68175b49-f1e6-4b34-aad9-20d61a43d427"
	B40ZhurTaaDruidOracle        = "5979310e-38c8-489f-ab9e-2723af98a3a5"
	B40ThreeTreeMascotOracle     = "6cb9d153-4eaa-4f32-8930-bd68cd99c128"
	B40JacesSanctumOracle        = "22d7048e-d538-4fd3-9fb3-2e99e74875fe"
	B40MechanizedWarfareOracle   = "c912a43b-8994-434e-84c0-f4cf58abbd42"
	B40ShivanDevastatorOracle    = "b7daa74c-6142-4107-9355-be98af6ccf13"
	B40BygoneColossusOracle      = "1bd584d5-4e11-428c-b51e-462e4292b07f"
	B40VraanOracle               = "b2f2645f-5f74-456a-bd02-83169d8b8a7e"
	B40HerdBalothOracle          = "03873314-6e64-43ab-95c0-3d8692a57a03"
	B40GixianPuppeteerOracle     = "9d6a9a37-a245-4503-baeb-9488553798ab"
	B40VeinwitchCovenOracle      = "ab478ac9-af59-4df1-afef-5e9806c06643"
	B40GargosOracle              = "378fcd0f-1096-4769-9149-55b4a889ff56"
	B40VincentOracle             = "b8961205-87fa-4cce-aba9-84fa2f91a67f"
	B40DazzlingDenialOracle      = "6d56bd32-f47a-4e54-a88e-b69d16283ea4"
	B40OliphauntOracle           = "186b2256-4af3-48cb-96b0-b0e80a7ee6dc"
	B40MurderousRedcapOracle     = "a498bc70-36e7-4454-bc44-906893df38b8"
	B40AngelOfInventionOracle    = "28c7f2b6-ba67-4ccf-9e1b-99c89a9d1f72"
	B40DragToTheRootsOracle      = "ef4c478c-7019-4ec0-8edf-2a3078a8e97a"
	B40DecoctionModuleOracle     = "1daba4f6-2a7d-426a-97d1-0298a7100c45"
	B40LumberingWorldwagonOracle = "7e60a641-1f2d-45ff-b3f2-d389be938b22"
)

// b40PricedFor runs one card through the cost pipeline as `caster`
// would pay it, and returns the total mana value they owe.
//
// Routed through ApplyCostModifiers rather than through a real cast,
// for the reason priceInHand is: the question is what the board
// charges, not whether the caster could satisfy timing and a pool.
func B40PricedFor(t *testing.T, g *game.Game, caster uuid.UUID, name, typeLine, manaCost string) int {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", manaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
			Owner: caster, Controller: caster,
		},
		Controller: caster,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers(%s): %v", name, err)
	}
	return out.ManaValue()
}

// b40PickCard answers a targeted trigger's pick_target prompt with a
// card — b16PickPlayer's sibling for the card slot.
func B40PickCard(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: card}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// b40TapForMana activates one named mana ability by index and answers
// the colour pick it opens — b28TapForMana with the index exposed,
// which Grand Coliseum's second ability and Three Tree Mascot both
// need.
func B40TapForMana(t *testing.T, g *game.Game, controller, card uuid.UUID, idx int, color string) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, idx, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility(%d): %v", idx, err)
	}
	if color != "" {
		B10ResolveAllManaPicks(t, g, controller, color)
	}
}
