package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// snow_cost_cards_test.go — the Kaldheim and Coldsnap cards that ride
// #5986's {S} payment (CR 107.4h). Each plays the real catalog entry and
// pays its {S} with a snow land, and the headline case is refused with
// nonsnow mana.

const (
	snowValkyrieOracle = "d42ecab7-f6ce-4e50-97de-7aa1ced64ebe"
	snowShadeOracle    = "1bbcfbad-c4c5-4818-b73c-26d83f7fa22c"
	snowDraugrOracle   = "793fe164-6c57-4fe0-a955-d5f7ffe388a9"
	snowRaptorOracle   = "9ca01d83-7794-4952-8cd8-c2ecb32cffd5"
	snowGriffinOracle  = "862ecb4e-7030-48d5-bbdf-655662059430"
)

func TestSnowCostCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		snowValkyrieOracle: "Hailstorm Valkyrie",
		snowShadeOracle:    "Chilling Shade",
		snowDraugrOracle:   "Grim Draugr",
		snowRaptorOracle:   "Frost Raptor",
		snowGriffinOracle:  "Boreal Griffin",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// snowLand puts a snow basic of `subtype` under `owner`'s control.
func snowLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return blbB1Push(g, owner, "Snow-Covered "+subtype, "Basic Snow Land — "+subtype, "", 0, 0)
}

// plainLand puts a nonsnow basic of `subtype` under `owner`'s control.
func plainLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return blbB1Push(g, owner, subtype, "Basic Land — "+subtype, "", 0, 0)
}

func landTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

func TestHailstormValkyriePaysSnowSnowWithSnowLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := blbB1Push(g, me.ID, "Hailstorm Valkyrie", "Snow Creature — Angel Wizard", snowValkyrieOracle, 2, 2)
	for _, kw := range []string{"flying", "trample"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("Hailstorm Valkyrie does not have %s", kw)
		}
	}
	advanceToMain(t, g)
	swamps := []uuid.UUID{plainLand(g, me.ID, "Swamp"), plainLand(g, me.ID, "Swamp")}

	// Two Swamps make two black mana, and no snow mana: refused, and
	// nothing tapped.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S}{S} auto-tapped from two nonsnow Swamps: accepted")
	}
	for _, s := range swamps {
		if landTapped(t, g, s) {
			t.Fatal("a refused {S}{S} tapped a Swamp")
		}
	}
	// Floating nonsnow mana is refused too.
	floatMana(t, g, me, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{S}{S} paid with floating nonsnow {B}{B}: accepted")
	}

	snow := []uuid.UUID{snowLand(g, me.ID, "Swamp"), snowLand(g, me.ID, "Island")}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("{S}{S} off a Snow-Covered Swamp and a Snow-Covered Island: %v", err)
	}
	for _, s := range snow {
		if !landTapped(t, g, s) {
			t.Error("the plan did not tap both snow lands")
		}
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("after the pump %d/%d, want 4/4", p, tough)
	}
}

func TestChillingShadePumpsOncePerSnowMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := blbB1Push(g, me.ID, "Chilling Shade", "Snow Creature — Shade", snowShadeOracle, 1, 1)
	advanceToMain(t, g)
	snowLand(g, me.ID, "Forest")
	snowLand(g, me.ID, "Mountain")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Error("a third {S} with both snow lands tapped: accepted")
	}
	if p, tough := ptOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("after two pumps %d/%d, want 3/3", p, tough)
	}
}

func TestGrimDraugrPaysItsGenericWithAnyManaAndItsSnowWithSnow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := blbB1Push(g, me.ID, "Grim Draugr", "Snow Creature — Zombie Berserker", snowDraugrOracle, 3, 2)
	advanceToMain(t, g)
	swamp := plainLand(g, me.ID, "Swamp")
	snow := snowLand(g, me.ID, "Swamp")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("{1}{S} off a Swamp and a Snow-Covered Swamp: %v", err)
	}
	if !landTapped(t, g, swamp) || !landTapped(t, g, snow) {
		t.Error("the plan did not tap both lands")
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 4 || tough != 2 {
		t.Errorf("after the pump %d/%d, want 4/2", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, id, "menace") {
		t.Error("Grim Draugr did not gain menace")
	}
}

func TestSnowKeywordGrantsPayWithSnowLands(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, granted string
		lands                           int
	}{
		{"Frost Raptor", "Snow Creature — Bird", snowRaptorOracle, "shroud", 2},
		{"Boreal Griffin", "Snow Creature — Griffin", snowGriffinOracle, "first strike", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := blbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !hasKeywordOnBattlefield(t, g, id, "flying") {
				t.Errorf("%s does not have flying", tc.name)
			}
			advanceToMain(t, g)
			plainLand(g, me.ID, "Plains")
			plainLand(g, me.ID, "Plains")
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
				t.Fatalf("%s's snow cost auto-tapped from nonsnow Plains: accepted", tc.name)
			}
			for i := 0; i < tc.lands; i++ {
				snowLand(g, me.ID, "Island")
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
				t.Fatalf("%s's snow cost off snow Islands: %v", tc.name, err)
			}
			passPriorityAroundTable(t, g)
			if !hasKeywordOnBattlefield(t, g, id, tc.granted) {
				t.Errorf("%s did not gain %s", tc.name, tc.granted)
			}
		})
	}
}
