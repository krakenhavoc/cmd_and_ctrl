package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_rest_view_test.go — the wire half of #789: the counter
// components on a MANA ability, the among and variable shapes, and the
// add-a-counter cost.
//
// The point of the embedded CounterCostView is that a mana ability and
// an activated ability ship the same field names from the same
// projection, so the client's one picker serves both. These tests are
// what stops that drifting back apart.

func seatVividLand(g *game.Game, owner uuid.UUID, charges int) uuid.UUID {
	id := uuid.New()
	c := game.Card{
		InstanceID: id,
		Name:       "Vivid Test Land",
		TypeLine:   "Land",
		OracleID:   "00000000-0000-0000-0000-0000000000v1",
		Owner:      owner,
		Controller: owner,
		ManaAbilities: []game.ManaAbilityShape{
			{TapCost: true, Produced: "{U}", Label: "Add {U}"},
			{
				TapCost:        true,
				RemoveCounters: &game.CounterRemovalCost{Counter: "charge", N: 1},
				Produced:       "{W|U|B|R|G}",
				Label:          "Add one mana of any color",
			},
		},
	}
	if charges > 0 {
		c.Counters = map[string]int{"charge": charges}
	}
	g.Battlefield.PushTop(c)
	return id
}

func TestManaAbilityViewCarriesItsCounterCost(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	land := seatVividLand(g, owner, 2)
	g.BumpLayerVersionForTest()

	card := vehicleView(t, g, land)
	if len(card.ManaAbilities) != 2 {
		t.Fatalf("mana abilities: %d, want 2", len(card.ManaAbilities))
	}
	plain, counter := card.ManaAbilities[0], card.ManaAbilities[1]
	if plain.CounterCostN != 0 {
		t.Errorf("the plain ability carries a counter cost: %+v", plain.CounterCostView)
	}
	if counter.CounterCostN != 1 {
		t.Errorf("counter_cost_n = %d, want 1", counter.CounterCostN)
	}
	if counter.CounterCostKind != "charge" {
		t.Errorf("counter_cost_kind = %q, want \"charge\"", counter.CounterCostKind)
	}
	if !counter.CounterCostSelf {
		t.Error("counter_cost_self is false for a \"from this land\" cost")
	}
	if len(counter.CounterCostOptions) != 1 || counter.CounterCostOptions[0].CardID != land.String() {
		t.Fatalf("options = %+v, want just the land itself", counter.CounterCostOptions)
	}
	if k := counter.CounterCostOptions[0].Kinds; len(k) != 1 || k[0].Kind != "charge" || k[0].Count != 2 {
		t.Errorf("kinds = %+v, want one charge entry counting 2", k)
	}
}

// With no charge counter left the option list is empty, which is what
// greys the client's row — the same fact the enumerator reads to stop
// offering the move.
func TestManaAbilityViewDropsTheOptionsWhenNothingCanPay(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	land := seatVividLand(g, owner, 0)
	g.BumpLayerVersionForTest()

	card := vehicleView(t, g, land)
	counter := card.ManaAbilities[1]
	if counter.CounterCostN != 1 {
		t.Errorf("counter_cost_n = %d, want 1 — the component is still declared", counter.CounterCostN)
	}
	if len(counter.CounterCostOptions) != 0 {
		t.Errorf("options = %+v, want none", counter.CounterCostOptions)
	}
}

// A variable removal ships its floor, its flag and the ceiling the
// picker's stepper needs — all from the walk the engine validates
// against, so a player cannot name a number the server refuses.
func TestManaAbilityViewCarriesTheVariableCeiling(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Mage-Ring Test Network", TypeLine: "Land",
		OracleID: "00000000-0000-0000-0000-0000000000v2",
		Owner:    owner, Controller: owner,
		Counters: map[string]int{"storage": 4},
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:        true,
			RemoveCounters: &game.CounterRemovalCost{Counter: "storage", Variable: true},
			Produced:       "{C}",
		}},
	})
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, id).ManaAbilities[0]
	if !ab.CounterCostVariable {
		t.Error("counter_cost_variable is false for an \"any number\" cost")
	}
	if ab.CounterCostMax != 4 {
		t.Errorf("counter_cost_max = %d, want 4", ab.CounterCostMax)
	}
}

// An among cost ships its flag and the whole pool it may draw from —
// the client sums them itself, because whether they add up to N is a
// question about the PAYMENT rather than about any one permanent.
func TestActivatedAbilityViewCarriesTheAmongPool(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatCounterCostSource(g, owner, &game.CounterRemovalCost{
		Counter: "+1/+1", N: 3, Among: true,
		From: &game.TargetSpec{
			Mode: "permanent", Label: "artifacts you control", Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsArtifact() },
			Min:    1, Max: 1,
		},
	})
	for _, n := range []int{1, 2} {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Artifact", TypeLine: "Artifact",
			Owner: owner, Controller: owner, Counters: map[string]int{"+1/+1": n},
		})
	}
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if !ab.CounterCostAmong {
		t.Error("counter_cost_among is false for a \"from among\" cost")
	}
	if ab.CounterCostN != 3 {
		t.Errorf("counter_cost_n = %d, want 3", ab.CounterCostN)
	}
	if len(ab.CounterCostOptions) != 2 {
		t.Fatalf("options = %+v, want both artifacts (one counter is a legal part)", ab.CounterCostOptions)
	}
}

// An add-a-counter cost ships its kind and count, and says whether CR
// 118.3 blocks it — the one reason such a cost is unpayable.
func TestActivatedAbilityViewCarriesTheAddCounterCost(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Devoted Test Druid", TypeLine: "Creature — Elf Druid",
		Owner: owner, Controller: owner, Power: 0, Toughness: 2,
		OracleID: "00000000-0000-0000-0000-0000000000de",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Untap this creature",
			Cost:  game.AbilityCost{AddCounter: &game.CounterAddCost{Counter: "-1/-1", N: 1}},
		}},
	})
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, id).ActivatedAbilities[0]
	if ab.CounterCostAdd != 1 || ab.CounterCostAddKind != "-1/-1" {
		t.Errorf("add cost = %d %q, want 1 \"-1/-1\"", ab.CounterCostAdd, ab.CounterCostAddKind)
	}
	if ab.CounterAddBlocked {
		t.Error("counter_add_blocked is true for a permanent that can have counters")
	}
}
