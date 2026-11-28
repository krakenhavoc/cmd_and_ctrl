package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b763_cards_test.go — Tempest lands, batch 763 (S82 older sets, #5628).
// Every test plays the real catalog entry through the engine.

type oldB763Land struct {
	name, oracle, a, b string
}

var (
	oldB763Painlands = []oldB763Land{
		{"Skyshroud Forest", "117a5fae-7fc4-4e24-b646-5727ea392fa7", "G", "U"},
		{"Scabland", "2896f01d-003b-4d68-9d0a-64990ba59cbe", "R", "W"},
		{"Pine Barrens", "603dc263-270c-4a16-9b75-41f36fd7dfde", "B", "G"},
		{"Salt Flats", "7a951bd7-4f7d-44ea-9c2f-fe2f6b2f5289", "W", "B"},
		{"Caldera Lake", "c737d27b-db14-4bd4-8f16-bcbd4401c47b", "U", "R"},
	}
	oldB763Freezers = []oldB763Land{
		{"Mogg Hollows", "1745fd57-467c-45f9-a46e-b9a2af87ec87", "R", "G"},
		{"Rootwater Depths", "2d28c83a-7415-4eb0-95a6-6245f2169d17", "U", "B"},
		{"Thalakos Lowlands", "5a54d6a3-b1d0-42fe-9531-604b34d197f1", "W", "U"},
		{"Cinder Marsh", "6f8cc374-e76c-4bfa-bf20-28dea0bfefbe", "B", "R"},
		{"Vec Townships", "b0a4680f-9707-431c-b5d5-7d4424783602", "G", "W"},
	}
	oldB763GhostTown = oldB763Land{"Ghost Town", "f2c861d3-b302-4e84-b647-099551007269", "", ""}
)

func TestOldB763EveryLandIsRegisteredFull(t *testing.T) {
	all := append(append([]oldB763Land{}, oldB763Painlands...), oldB763Freezers...)
	all = append(all, oldB763GhostTown)
	for _, l := range all {
		spec, ok := effects.Lookup(l.oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", l.name)
			continue
		}
		if spec.Name != l.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v with %d caveats, want Full with none", l.name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB763Pool(me *game.Player) string {
	out := ""
	for _, u := range me.ManaPool.Units() {
		out += u.Color
	}
	return out
}

func TestOldB763TappedPainlandsEnterTappedAndPainOnlyTheColouredHalf(t *testing.T) {
	for _, l := range oldB763Painlands {
		t.Run(l.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.PlayLandFromHand(t, g, l.name, l.oracle)
			if tapped, ok := effectstest.BattlefieldCardTapped(g, land); !ok || !tapped {
				t.Fatalf("the land entered untapped (on battlefield %v)", ok)
			}
			effectstest.RtUntap(g, land)
			life := me.Life
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("colourless half: %v", err)
			}
			if oldB763Pool(me) != "C" || me.Life != life {
				t.Errorf("colourless half gave %q and life %d→%d, want {C} and no damage", oldB763Pool(me), life, me.Life)
			}
			effectstest.RtUntap(g, land)
			me.ManaPool.EmptyPool()
			if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{l.a}}); err != nil {
				t.Fatalf("coloured half: %v", err)
			}
			if me.Life != life-1 || oldB763Pool(me) != l.a {
				t.Errorf("coloured half gave %q and life %d→%d, want {%s} and 1 damage", oldB763Pool(me), life, me.Life, l.a)
			}
		})
	}
}

func TestOldB763FreezingLandsSitOutTheNextUntapStepOnlyAfterTheColouredHalf(t *testing.T) {
	for _, l := range oldB763Freezers {
		t.Run(l.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			colourless := effectstest.SeedPermanentWithOracle(g, me.ID, l.name, "Land", l.oracle)
			coloured := effectstest.SeedPermanentWithOracle(g, me.ID, l.name, "Land", l.oracle)
			if err := g.ActivateManaAbility(me.ID, colourless, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("colourless half: %v", err)
			}
			if err := g.ActivateManaAbility(me.ID, coloured, 1, game.ManaAbilityParams{Colors: []string{l.b}}); err != nil {
				t.Fatalf("coloured half: %v", err)
			}
			if !contains763(oldB763Pool(me), l.b) {
				t.Errorf("the coloured half made %q, want {%s} among it", oldB763Pool(me), l.b)
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.AdvanceToUpkeepOf(t, g, 0)
			if tapped, _ := effectstest.BattlefieldCardTapped(g, colourless); tapped {
				t.Error("the land tapped for {C} did not untap")
			}
			if tapped, _ := effectstest.BattlefieldCardTapped(g, coloured); !tapped {
				t.Error("the land tapped for colour untapped during the next untap step")
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			effectstest.AdvanceToUpkeepOf(t, g, 0)
			if tapped, _ := effectstest.BattlefieldCardTapped(g, coloured); tapped {
				t.Error("the land stayed tapped through a second untap step")
			}
		})
	}
}

func contains763(pool, colour string) bool {
	for _, r := range pool {
		if string(r) == colour {
			return true
		}
	}
	return false
}

func TestOldB763GhostTownReturnsToHandOnlyOnSomeoneElsesTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	town := effectstest.SeedPermanentWithOracle(g, me.ID, "Ghost Town", "Land", oldB763GhostTown.oracle)
	if err := g.ActivateManaAbility(me.ID, town, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if oldB763Pool(me) != "C" {
		t.Errorf("Ghost Town made %q, want {C}", oldB763Pool(me))
	}
	if err := g.ActivateCatalogAbility(me.ID, town, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("returned it to hand on its controller's own turn")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if err := g.ActivateCatalogAbility(me.ID, town, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate on an opponent's turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(town) {
		t.Error("Ghost Town stayed on the battlefield")
	}
	if !oldB763InHand(me, "Ghost Town") {
		t.Error("Ghost Town is not in its owner's hand")
	}
}

func oldB763InHand(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

var _ = uuid.Nil
