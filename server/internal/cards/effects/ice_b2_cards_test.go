package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ice_b2_cards_test.go — Ice Age, batch 2 (#5636). Each test plays the
// real catalog entry and asserts what its text says. The seeding and
// activation helpers are batch 1's (ice_b1_cards_test.go).

const (
	iceB2BalduvianBearsOracle    = "32c10a3e-85b0-49af-9b23-81d0b7e407df"
	iceB2ScaledWurmOracle        = "009a8e38-74d0-4b4a-b8ca-62c9f3183531"
	iceB2TorGiantOracle          = "2457a9f7-8a46-4c7a-b35e-bcbf4a79c365"
	iceB2SilverErneOracle        = "6ca90451-861d-4aa8-93e7-3f3e02385530"
	iceB2WindSpiritOracle        = "d4c22b68-c4cb-447a-97df-431e2f1e31f2"
	iceB2GlacialWallOracle       = "7fcf8faa-2c93-4dc8-aa56-6fc40340e212"
	iceB2MoorFiendOracle         = "ff1fd718-ee97-425c-b8bf-97fea192c631"
	iceB2PygmyAllosaurusOracle   = "8c4b1ab2-01d9-486a-a9db-ac7c72500d30"
	iceB2MountainGoatOracle      = "031cd561-d9b1-4a0e-bc38-bc09e80563c3"
	iceB2PaleBearsOracle         = "0acb070b-6e81-4e0e-925d-14da9b931355"
	iceB2FlameSpiritOracle       = "123bfa86-4265-448f-97cd-2c0614212862"
	iceB2SeaSpiritOracle         = "0ebd8ab3-d4b7-480e-b7b1-267078edddf2"
	iceB2FolkOfThePinesOracle    = "6cd8cf2a-972f-41b7-ab0c-b4953accc2fe"
	iceB2ShamblingStriderOracle  = "fd65468c-d75a-4be2-8a0c-9c4e808e3dd0"
	iceB2WallOfPineNeedlesOracle = "1d5bc31c-f4a1-4a59-b466-0347ddbbc7fb"
	iceB2YavimayaGnatsOracle     = "828c1a94-36e1-484b-afc8-87f85956d60d"
)

func TestIceB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		iceB2BalduvianBearsOracle:    "Balduvian Bears",
		iceB2ScaledWurmOracle:        "Scaled Wurm",
		iceB2TorGiantOracle:          "Tor Giant",
		iceB2SilverErneOracle:        "Silver Erne",
		iceB2WindSpiritOracle:        "Wind Spirit",
		iceB2GlacialWallOracle:       "Glacial Wall",
		iceB2MoorFiendOracle:         "Moor Fiend",
		iceB2PygmyAllosaurusOracle:   "Pygmy Allosaurus",
		iceB2MountainGoatOracle:      "Mountain Goat",
		iceB2PaleBearsOracle:         "Pale Bears",
		iceB2FlameSpiritOracle:       "Flame Spirit",
		iceB2SeaSpiritOracle:         "Sea Spirit",
		iceB2FolkOfThePinesOracle:    "Folk of the Pines",
		iceB2ShamblingStriderOracle:  "Shambling Strider",
		iceB2WallOfPineNeedlesOracle: "Wall of Pine Needles",
		iceB2YavimayaGnatsOracle:     "Yavimaya Gnats",
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

// --- vanilla creatures -------------------------------------------------------------

func TestIceB2VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for oracle, name := range map[string]string{
		iceB2BalduvianBearsOracle: "Balduvian Bears",
		iceB2ScaledWurmOracle:     "Scaled Wurm",
		iceB2TorGiantOracle:       "Tor Giant",
	} {
		spec, _ := Lookup(oracle)
		if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Triggered) != 0 ||
			len(spec.Static) != 0 || len(spec.BlockRules) != 0 {
			t.Errorf("%s is vanilla but its spec carries abilities: %+v", name, spec)
		}
	}
}

// --- printed keywords --------------------------------------------------------------

func TestIceB2CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Silver Erne", iceB2SilverErneOracle, []string{"flying", "trample"}},
		{"Wind Spirit", iceB2WindSpiritOracle, []string{"flying", "menace"}},
		{"Glacial Wall", iceB2GlacialWallOracle, []string{"defender"}},
		{"Moor Fiend", iceB2MoorFiendOracle, []string{"swampwalk"}},
		{"Pygmy Allosaurus", iceB2PygmyAllosaurusOracle, []string{"swampwalk"}},
		{"Mountain Goat", iceB2MountainGoatOracle, []string{"mountainwalk"}},
		{"Pale Bears", iceB2PaleBearsOracle, []string{"islandwalk"}},
		{"Wall of Pine Needles", iceB2WallOfPineNeedlesOracle, []string{"defender"}},
		{"Yavimaya Gnats", iceB2YavimayaGnatsOracle, []string{"flying"}},
	} {
		g := newCatalogGame(t)
		id := iceB1Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

// --- landwalk ------------------------------------------------------------------------

func TestIceB2LandwalkersCantBeBlockedOnlyWhileTheDefenderHasTheLand(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, land, other string
	}{
		{"Moor Fiend", iceB2MoorFiendOracle, "Swamp", "Forest"},
		{"Pygmy Allosaurus", iceB2PygmyAllosaurusOracle, "Swamp", "Island"},
		{"Mountain Goat", iceB2MountainGoatOracle, "Mountain", "Plains"},
		{"Pale Bears", iceB2PaleBearsOracle, "Island", "Swamp"},
	} {
		for _, defenderHasIt := range []bool{true, false} {
			land := tc.other
			if defenderHasIt {
				land = tc.land
			}
			t.Run(tc.name+" against a "+land, func(t *testing.T) {
				g := newCatalogGame(t)
				me, them := seatsForCombat(g)
				walker := iceB1Push(g, me.ID, tc.name, "Creature", tc.oracle, 2, 2)
				bear := iceB1Push(g, them.ID, "Bear", "Creature — Bear", "", 2, 2)
				seedLand(g, them.ID, land, "Basic Land — "+land, "")
				// The attacker's own land of the type does not count: only
				// the defending player's does.
				seedLand(g, me.ID, tc.land, "Basic Land — "+tc.land, "")
				attackIntoBlocks(t, g, them.ID, walker)
				r := blockRefusal(t, g, bear, walker)
				switch {
				case defenderHasIt && r != game.BlockReasonLandwalk:
					t.Errorf("blocking %s while the defender controls a %s: reason %q, want landwalk", tc.name, land, r)
				case !defenderHasIt && r != "":
					t.Errorf("blocking %s while the defender controls only a %s was refused: %q", tc.name, land, r)
				}
			})
		}
	}
}

// --- the pumps ---------------------------------------------------------------------

func TestIceB2SelfPumpsGrowTheirSourceEachActivation(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		p, tough                     int
		dp, dt                       int
	}{
		{"Flame Spirit", "Creature — Elemental Spirit", iceB2FlameSpiritOracle, "{R}", 2, 3, 1, 0},
		{"Sea Spirit", "Creature — Elemental Spirit", iceB2SeaSpiritOracle, "{U}", 2, 3, 1, 0},
		{"Folk of the Pines", "Creature — Dryad", iceB2FolkOfThePinesOracle, "{C}{G}", 2, 5, 1, 0},
		{"Shambling Strider", "Creature — Yeti", iceB2ShamblingStriderOracle, "{R}{G}", 5, 5, 1, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := iceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			for i := 0; i < 2; i++ {
				dcMana(t, g, me.ID, tc.mana)
				if err := iceB1Activate(g, me.ID, id, 0); err != nil {
					t.Fatalf("activation %d: %v", i+1, err)
				}
				passPriorityAroundTable(t, g)
			}
			if p, tough := ptOf(t, g, id); p != tc.p+2*tc.dp || tough != tc.tough+2*tc.dt {
				t.Errorf("%s is %d/%d after two pumps, want %d/%d", tc.name, p, tough, tc.p+2*tc.dp, tc.tough+2*tc.dt)
			}
			if err := iceB1Activate(g, me.ID, id, 0); err == nil {
				t.Error("the pump was activated with an empty pool")
			}
		})
	}
}

// Shambling Strider's -1 is real toughness loss: enough activations
// kill it through the state-based actions.
func TestIceB2ShamblingStriderDiesAtZeroToughness(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := iceB1Push(g, me.ID, "Shambling Strider", "Creature — Yeti", iceB2ShamblingStriderOracle, 5, 5)
	for i := 0; i < 5; i++ {
		dcMana(t, g, me.ID, "{R}{G}")
		if err := iceB1Activate(g, me.ID, id, 0); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Fatal("Shambling Strider survived at 0 toughness")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("Shambling Strider is not in its owner's graveyard")
	}
}

// --- regeneration ------------------------------------------------------------------

func TestIceB2RegeneratorsSurviveTheirNextDestruction(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Wall of Pine Needles", "Creature — Plant Wall", iceB2WallOfPineNeedlesOracle, 3, 3},
		{"Yavimaya Gnats", "Creature — Insect", iceB2YavimayaGnatsOracle, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := iceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			if err := iceB1Activate(g, me.ID, id, 0); err == nil {
				t.Fatal("the regeneration was activated with an empty pool")
			}
			dcMana(t, g, me.ID, "{G}")
			if err := iceB1Activate(g, me.ID, id, 0); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if n := regenShieldsOn(t, g, id); n != 1 {
				t.Fatalf("%d regeneration shields, want 1", n)
			}
			destroyIt(t, g, id)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s was destroyed through its shield", tc.name)
			}
			if !findCardAnywhere(t, g, id).Tapped {
				t.Error("the regenerated creature is untapped")
			}
			destroyIt(t, g, id)
			if g.Battlefield.Contains(id) {
				t.Errorf("one shield saved %s twice", tc.name)
			}
		})
	}
}
