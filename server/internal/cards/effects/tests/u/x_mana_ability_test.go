package testsu

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// x_mana_ability_test.go — #6125: X in a mana ability's cost. A mana
// ability is an activated ability (CR 605.1a) activated by CR 602.2b's
// process, so its controller announces X with it (CR 107.3a) and it
// resolves at once (CR 605.3b).
//
//   - "{X}, {T}, Sacrifice this artifact: Add X mana in any combination
//     of colors." (Wizard's Rockets)
//   - "{T}, Sacrifice X Goats: Add X mana of any one color. You gain X
//     life." (Springjack Pasture)

const (
	xmaRocketsOracle = "00000000-6125-4000-8000-000000000001"
	xmaPastureOracle = "00000000-6125-4000-8000-000000000002"
)

func registerXmaRockets(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     xmaRocketsOracle,
		Name:         "Test Rockets",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:            effects.ManaAbilityCost{Mana: "{X}", Tap: true, Sacrifice: true},
			ProducedForPaid: effects.XManaAnyCombination,
			Label:           "{X}, {T}, Sacrifice this artifact: Add X mana in any combination of colors.",
		}},
	})
}

func registerXmaPasture(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     xmaPastureOracle,
		Name:         "Test Pasture",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:            effects.ManaAbilityCost{Tap: true, SacrificeOther: effects.SacrificeX("X Goats", effects.HasSubtype("Goat")).SacrificeOther},
			ProducedForPaid: effects.XManaOfAnyOneColor,
			GainLifeX:       true,
			Label:           "{T}, Sacrifice X Goats: Add X mana of any one color. You gain X life.",
		}},
	})
}

func xmaPool(me *game.Player) map[string]int {
	out := map[string]int{}
	for _, tok := range me.ManaPool {
		out[tok.Color] += tok.N()
	}
	return out
}

// X = 2: two generic mana pay the {X}, the Rockets are sacrificed, and
// two mana of the named colours are added.
func TestXManaAbilityPaysXAndAddsXInAnyCombination(t *testing.T) {
	registerXmaRockets(t)
	g, me, _ := effectstest.P7Table(t)
	rockets := effectstest.PushCatalogPermanent(g, me.ID, "Test Rockets", "Artifact", xmaRocketsOracle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateManaAbility(me.ID, rockets, 0, game.ManaAbilityParams{XValue: 2, Colors: []string{"R", "G"}}); err != nil {
		t.Fatalf("activate X=2: %v", err)
	}
	pool := xmaPool(me)
	if pool["R"] != 1 || pool["G"] != 1 || pool["C"] != 0 {
		t.Fatalf("pool %v, want one R and one G (the {C}{C} spent on X)", pool)
	}
	if _, ok := effectstest.BattlefieldCard(g, rockets); ok {
		t.Fatal("the Rockets were not sacrificed")
	}
}

// X the pool cannot pay is refused with nothing spent; an X on an
// ability with no X is refused; a negative X is refused.
func TestXManaAbilityRefusals(t *testing.T) {
	registerXmaRockets(t)
	g, me, _ := effectstest.P7Table(t)
	rockets := effectstest.PushCatalogPermanent(g, me.ID, "Test Rockets", "Artifact", xmaRocketsOracle, false)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateManaAbility(me.ID, rockets, 0, game.ManaAbilityParams{XValue: 2}); err == nil {
		t.Fatal("X=2 with one mana floating was accepted")
	}
	if err := g.ActivateManaAbility(me.ID, rockets, 0, game.ManaAbilityParams{XValue: -1}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("X=-1: err %v, want ErrInvalidParam", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, rockets); !ok || xmaPool(me)["C"] != 1 {
		t.Fatal("a refused activation moved the Rockets or spent mana")
	}
	forest := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{XValue: 1}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("X on a Forest: err %v, want ErrInvalidParam", err)
	}
}

// Sacrifice X Goats with X = 2: two Goats go, two mana of the one named
// colour are added, and the controller gains 2 life. The count must be X.
func TestSacrificeXGoatsAddsXOfOneColorAndGainsXLife(t *testing.T) {
	registerXmaPasture(t)
	g, me, _ := effectstest.P7Table(t)
	pasture := effectstest.PushCatalogPermanent(g, me.ID, "Test Pasture", "Land", xmaPastureOracle, false)
	goat := func() uuid.UUID {
		return effectstest.PushPermanent(g, me.ID, game.Card{Name: "Goat", TypeLine: "Token Creature — Goat", Power: 0, Toughness: 1})
	}
	a, b, c := goat(), goat(), goat()
	if err := g.ActivateManaAbility(me.ID, pasture, 0, game.ManaAbilityParams{XValue: 2, SacrificeIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("one Goat for X=2 was accepted")
	}
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, pasture, 0, game.ManaAbilityParams{XValue: 2, SacrificeIDs: []uuid.UUID{a, b}, Colors: []string{"B"}}); err != nil {
		t.Fatalf("X=2: %v", err)
	}
	if pool := xmaPool(me); pool["B"] != 2 || len(pool) != 1 {
		t.Fatalf("pool %v, want {B}{B}", pool)
	}
	if me.Life != life+2 {
		t.Fatalf("life %d -> %d, want +2", life, me.Life)
	}
	if _, ok := effectstest.BattlefieldCard(g, c); !ok {
		t.Fatal("the third Goat was sacrificed")
	}
}

// The enumerator offers each X ability with an X its payment matches, and
// the auto-tapper never plans one.
func TestXManaAbilityMovesAnnounceTheirX(t *testing.T) {
	registerXmaRockets(t)
	registerXmaPasture(t)
	g, me, _ := effectstest.P7Table(t)
	rockets := effectstest.PushCatalogPermanent(g, me.ID, "Test Rockets", "Artifact", xmaRocketsOracle, false)
	pasture := effectstest.PushCatalogPermanent(g, me.ID, "Test Pasture", "Land", xmaPastureOracle, false)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Goat", TypeLine: "Token Creature — Goat", Power: 0, Toughness: 1})
	effectstest.FloatMana(t, g, me, "{C}")
	seen := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeActivateManaAbility || (m.Source != rockets && m.Source != pasture) {
			continue
		}
		var p struct {
			XValue       int      `json:"x_value"`
			SacrificeIDs []string `json:"sacrifice_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		seen[m.Source] = true
		if m.Source == pasture && p.XValue != len(p.SacrificeIDs) {
			t.Errorf("pasture move X=%d sacrifices %d", p.XValue, len(p.SacrificeIDs))
		}
		if m.Source == rockets && p.XValue < 1 {
			t.Errorf("rockets move announces X=%d", p.XValue)
		}
	}
	if !seen[rockets] || !seen[pasture] {
		t.Fatalf("moves offered: %v", seen)
	}
	cost, _ := game.ParseCost("{2}")
	if plan, _ := g.AutoTapForCost(me.ID, cost, 0); len(plan) > 0 {
		for _, id := range plan {
			if id == rockets || id == pasture {
				t.Fatalf("the auto-tapper planned an X ability: %v", plan)
			}
		}
	}
}
