package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_view_test.go — the wire half of #625's counter-removal
// cost. The client needs the component's shape (how many, which kind,
// from what) and the options that can pay right now, and the options
// have to be the NON-targeting candidate set: a hexproof planeswalker
// pays Heart of Kiran's crew exactly as it pays in the engine.

func seatCounterCostSource(g *game.Game, owner uuid.UUID, rc *game.CounterRemovalCost) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Counter Cost Source",
		TypeLine:   "Artifact — Vehicle",
		OracleID:   "00000000-0000-0000-0000-0000000000dd",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "Crew — remove a loyalty counter", Cost: game.AbilityCost{RemoveCounters: rc}},
		},
	})
	return id
}

func seatCostWalker(g *game.Game, controller uuid.UUID, loyalty int, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: controller, Controller: controller, Keywords: keywords,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func planeswalkerCostSpec() *game.TargetSpec {
	return &game.TargetSpec{
		Mode: "permanent", Label: "a planeswalker you control", Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsPlaneswalker() },
		Min:    1, Max: 1,
	}
}

func TestActivatedAbilityViewCarriesCounterCostAndOptions(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatCounterCostSource(g, owner, &game.CounterRemovalCost{
		Counter: game.CounterLoyalty, N: 1, From: planeswalkerCostSpec(),
	})
	small := seatCostWalker(g, owner, 1)
	// Hexproof and shroud: untargetable, and still a legal payment —
	// choosing a permanent to pay a cost does not target it.
	big := seatCostWalker(g, owner, 5, "hexproof", "shroud")
	seatCostWalker(g, owner, 0)         // no loyalty: cannot pay
	seatCostWalker(g, g.Seats[1].ID, 7) // an opponent's: cannot pay
	g.Battlefield.PushTop(game.Card{    // loyalty counters on a non-walker: cannot pay
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.CounterCostN != 1 || ab.CounterCostKind != game.CounterLoyalty || ab.CounterCostSelf {
		t.Errorf("counter cost shape: n=%d kind=%q self=%v", ab.CounterCostN, ab.CounterCostKind, ab.CounterCostSelf)
	}
	if ab.CounterCostLabel != "a planeswalker you control" {
		t.Errorf("counter_cost_label = %q", ab.CounterCostLabel)
	}
	if len(ab.CounterCostOptions) != 2 {
		t.Fatalf("counter_cost_options = %+v, want the two payable walkers of mine", ab.CounterCostOptions)
	}
	if ab.CounterCostOptions[0].CardID != big.String() || ab.CounterCostOptions[1].CardID != small.String() {
		t.Errorf("options not most-loyalty-first, or the hexproof walker is missing: %+v", ab.CounterCostOptions)
	}
	if k := ab.CounterCostOptions[0].Kinds; len(k) != 1 || k[0].Kind != game.CounterLoyalty || k[0].Count != 5 {
		t.Errorf("kinds on the big walker = %+v, want loyalty ×5", k)
	}
	if ab.SorcerySpeed || ab.LoyaltyCost != nil {
		t.Error("a counter cost is not a loyalty ability and is not sorcery speed")
	}
}

// An empty pool is absent, which is what the client greys the row on,
// and the self form says so and needs no label.
func TestCounterCostViewSelfFormAndEmptyPool(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatCounterCostSource(g, owner, &game.CounterRemovalCost{Counter: "gold", N: 1})

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if !ab.CounterCostSelf || ab.CounterCostLabel != "" || ab.CounterCostKind != "gold" {
		t.Errorf("self form: self=%v label=%q kind=%q", ab.CounterCostSelf, ab.CounterCostLabel, ab.CounterCostKind)
	}
	if len(ab.CounterCostOptions) != 0 {
		t.Errorf("no gold counters, options %+v", ab.CounterCostOptions)
	}
	raw, _ := json.Marshal(ab)
	if strings.Contains(string(raw), "counter_cost_options") {
		t.Errorf("an empty pool should be absent on the wire: %s", raw)
	}

	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == src {
			g.Battlefield.Cards[i].Counters = map[string]int{"gold": 2}
		}
	}
	ab = vehicleView(t, g, src).ActivatedAbilities[0]
	if len(ab.CounterCostOptions) != 1 || ab.CounterCostOptions[0].CardID != src.String() {
		t.Errorf("self form with gold: options %+v, want the source", ab.CounterCostOptions)
	}
}

// "A counter" of any kind ships no kind and lists every kind each
// permanent could pay with.
func TestCounterCostViewAnyKindListsKinds(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatCounterCostSource(g, owner, &game.CounterRemovalCost{N: 1, From: &game.TargetSpec{
		Mode: "permanent", Label: "a creature you control", Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsCreature() },
	}})
	bear := seatCreature(g, owner, "Bear", 2, false, false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bear {
			g.Battlefield.Cards[i].Counters = map[string]int{game.CounterStun: 1, game.CounterPlusOne: 3}
		}
	}
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.CounterCostKind != "" {
		t.Errorf("any-kind cost shipped kind %q", ab.CounterCostKind)
	}
	if len(ab.CounterCostOptions) != 1 {
		t.Fatalf("options %+v, want the bear", ab.CounterCostOptions)
	}
	k := ab.CounterCostOptions[0].Kinds
	if len(k) != 2 || k[0].Kind != game.CounterPlusOne || k[1].Kind != game.CounterStun {
		t.Errorf("kinds %+v, want +1/+1 (3) then stun (1)", k)
	}
}

func TestNonCounterCostAbilityCarriesNoCounterFields(t *testing.T) {
	g := buildActiveGame(t)
	vehicle, _ := seatCrewVehicle(g)
	ab := vehicleView(t, g, vehicle).ActivatedAbilities[0]
	if ab.CounterCostN != 0 || ab.CounterCostOptions != nil || ab.CounterCostSelf || ab.CounterCostKind != "" {
		t.Errorf("a crew ability carried counter-cost fields: %+v", ab)
	}
}
