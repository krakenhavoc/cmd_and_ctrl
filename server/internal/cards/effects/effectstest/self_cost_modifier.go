package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// self_cost_modifier_test.go — #746: the cards that change their OWN
// cost (ADR 0048 addendum). The engine's ordering, floor, zone, face
// and target rules are pinned in game/self_cost_modifier_test.go; these
// tests pin the DECLARATIONS — that each card counts the right things
// for the right player — and the effects that ship with them.

const (
	ThoughtMonitorOracle    = "9deded8b-cec4-4ede-a50b-131404d456d4"
	MyrEnforcerOracle       = "2d8e1054-654f-42e8-8c29-12c3cf13a3eb"
	ThoughtcastOracle       = "cce9bbff-82dc-4b2f-addd-d6715588de20"
	VoyageHomeOracle        = "5c0cbc44-6c31-44fe-a3da-a97a37a01726"
	GhaltaPrimalOracle      = "b0b6be0c-41cf-4757-9f0e-87227b6ba6b3"
	VanquishHordeOracle     = "a332e80a-dc51-4dc6-bc85-e114a1c6fdb8"
	FireballOracle          = "aa7714b0-2bfb-458a-8ebf-37ec2c53383e"
	CallCoppercoatsOracle   = "73d8c33d-916a-4220-96ae-9622aad36210"
	WingedWordsOracle       = "c623aeb1-e6d4-48fe-bd2a-a7a6729aa4df"
	WizardsRetortOracle     = "b828251c-86a9-454f-9852-d0876d0f5153"
	IntoTheStoryOracle      = "f290c2e4-ab52-44d4-bdeb-31aeb835b18d"
	HourOfRevelationOracle  = "920c1dd4-c3f6-4020-a6f0-2e8acad2c212"
	MortalitySpearOracle    = "544acde0-850d-4c4f-8389-e22930d87345"
	CurtainsCallOracle      = "cfb1b64b-bf78-4015-ac29-cdcdf2adaa02"
	PriceOfFameOracle       = "45d5cd4c-7285-4507-8cbf-eace7a734f41"
	ShadowOfMortalityOracle = "45378874-aa51-4bb6-a161-be336c164778"
	WitherbloomOracle       = "fbb04a21-e513-4317-aac8-fa6df91c3438"
	MycosynthGolemOracle    = "ebcd864a-b7dd-4330-89e1-80576a9437ec"
	HamzaOracle             = "0cde6a29-d597-4feb-8b2b-528d2d57a1f3"
	AncientStoneIdolOracle  = "5f981cca-5cd9-49e4-ab1a-bbbf6fc7e737"
	HagraMaulingOracle      = "37783ce6-af58-4ef6-8ab4-587079970307"
)

// selfPriced is what the board charges `seat` to cast a card carrying
// `oracle` from `from` with `targets`, through the public pricing
// surface the auto-tap preview uses.
func SelfPriced(t *testing.T, g *game.Game, seat *game.Player, oracle, typeLine, manaCost string, from game.ZoneKind, targets []game.TargetRef) game.ParsedCost {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", manaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: "priced", TypeLine: typeLine, ManaCost: manaCost,
			OracleID: oracle, Owner: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   from,
		Targets:    targets,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers(%s): %v", oracle, err)
	}
	return out
}

// selfPricedMV is selfPriced from hand with no targets, as a mana value.
func SelfPricedMV(t *testing.T, g *game.Game, seat *game.Player, oracle, typeLine, manaCost string) int {
	t.Helper()
	return SelfPriced(t, g, seat, oracle, typeLine, manaCost, game.ZoneHand, nil).ManaValue()
}

// pushPermanent drops a plain permanent on the battlefield.
func PushPermanent(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	g.Battlefield.PushTop(c)
	return c.InstanceID
}
