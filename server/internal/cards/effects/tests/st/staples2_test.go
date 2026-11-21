package testsst

// staples2_test.go — second Commander-staples batch: sacrifice-for-land
// ramp, the disenchant/removal spread, and Merciless Eviction. As in
// staples_test.go, each test pins the property that makes the card
// worth playing rather than merely that it resolved.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- ramp spells ------------------------------------------------

func TestThreeVisitsFetchesForestUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	forestID := effectstest.SeedBasicsInLibrary(g.Seats[0], 1)[0]

	effectstest.CastCatalogSpell(t, g, "Three Visits", "Sorcery", effectstest.ThreeVisitsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	got, ok := effectstest.BattlefieldCard(g, forestID)
	if !ok {
		t.Fatal("Forest not fetched")
	}
	if got.Tapped {
		t.Error("Three Visits' land must enter untapped")
	}
}

// Two lands, both untapped — the reason Skyshroud Claim beats
// Explosive Vegetation when a deck can cast it.
func TestSkyshroudClaimFetchesTwoUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ids := effectstest.SeedBasicsInLibrary(g.Seats[0], 2)

	effectstest.CastCatalogSpell(t, g, "Skyshroud Claim", "Sorcery", effectstest.SkyshroudClaimOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	for _, id := range ids {
		got, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("Forest %s not fetched", id)
		}
		if got.Tapped {
			t.Error("Skyshroud Claim's lands must enter untapped")
		}
	}
}

func TestExplosiveVegetationFetchesTwoTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ids := effectstest.SeedBasicsInLibrary(g.Seats[0], 2)

	effectstest.CastCatalogSpell(t, g, "Explosive Vegetation", "Sorcery", effectstest.ExplosiveVegetationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	for _, id := range ids {
		got, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("basic %s not fetched", id)
		}
		if !got.Tapped {
			t.Error("Explosive Vegetation's lands must enter tapped")
		}
	}
}

// --- sacrifice-for-land ability shapes --------------------------

// These three differ only in their cost components, and the cost is
// the whole difference between them, so the shapes are pinned
// directly. No tap on the Elder or the Hart means both can be cashed
// in the turn they arrive; the Bauble's tap means it can't be used
// through a tapper.
func TestSacrificeForLandCostShapes(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana string
		tap                bool
	}{
		{"Sakura-Tribe Elder", effectstest.SakuraTribeElderOracle, "", false},
		{"Burnished Hart", effectstest.BurnishedHartOracle, "{3}", false},
		{"Wayfarer's Bauble", effectstest.WayfarersBaubleOracle, "{2}", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
			if len(abilities) != 1 {
				t.Fatalf("%d activated abilities, want 1", len(abilities))
			}
			cost := abilities[0].Cost
			if !cost.SacrificeSelf {
				t.Error("cost must sacrifice the source")
			}
			if cost.Tap != tc.tap {
				t.Errorf("Tap = %v, want %v", cost.Tap, tc.tap)
			}
			if cost.Mana != tc.mana {
				t.Errorf("Mana = %q, want %q", cost.Mana, tc.mana)
			}
			if abilities[0].Targets != nil {
				t.Error("fetching a land targets nothing")
			}
		})
	}
}

func TestPalladiumMyrManaAbility(t *testing.T) {
	abilities := game.ManaAbilitiesForCard(game.Card{OracleID: effectstest.PalladiumMyrOracle})
	if len(abilities) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(abilities))
	}
	if !abilities[0].TapCost {
		t.Error("cost should be a tap")
	}
	if abilities[0].Produced != "{C}{C}" {
		t.Errorf("produced %q, want {C}{C}", abilities[0].Produced)
	}
}

// --- removal spread ---------------------------------------------

// One table for the four "destroy/exile a permanent matching a type
// pair" spells: the interesting part is which types each accepts and
// which it must refuse.
func TestRemovalSpreadHitsAndMisses(t *testing.T) {
	type target struct {
		name, typeLine, manaCost string
		legal                    bool
	}
	for _, tc := range []struct {
		card, oracle, typeLine string
		targets                []target
	}{
		{"Naturalize", effectstest.NaturalizeOracle, "Instant", []target{
			{"Sol Ring", "Artifact", "{1}", true},
			{"Rhystic Study", "Enchantment", "{3}{U}", true},
			{"Grizzly Bears", "Creature — Bear", "{1}{G}", false},
		}},
		{"Mortify", effectstest.MortifyOracle, "Instant", []target{
			{"Grizzly Bears", "Creature — Bear", "{1}{G}", true},
			{"Rhystic Study", "Enchantment", "{3}{U}", true},
			{"Sol Ring", "Artifact", "{1}", false},
		}},
		{"Putrefy", effectstest.PutrefyOracle, "Instant", []target{
			{"Sol Ring", "Artifact", "{1}", true},
			{"Grizzly Bears", "Creature — Bear", "{1}{G}", true},
			{"Rhystic Study", "Enchantment", "{3}{U}", false},
		}},
	} {
		for _, tgt := range tc.targets {
			t.Run(tc.card+"/"+tgt.typeLine, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				victim := g.Seats[1]
				id := uuid.New()
				g.Battlefield.PushTop(game.Card{
					InstanceID: id, Name: tgt.name, TypeLine: tgt.typeLine,
					ManaCost: tgt.manaCost, Power: 2, Toughness: 2,
					Owner: victim.ID, Controller: victim.ID,
				})

				active := g.Seats[g.Turn.ActiveSeat]
				spellID := uuid.New()
				active.Hand.PushTop(game.Card{
					InstanceID: spellID, Name: tc.card, TypeLine: tc.typeLine,
					OracleID: tc.oracle, Owner: active.ID, Controller: active.ID,
				})
				for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
					if _, err := g.AdvanceStep(); err != nil {
						t.Fatalf("AdvanceStep: %v", err)
					}
				}
				err := g.CastSpell(active.ID, spellID, game.CastSpellParams{
					Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
				})
				if tgt.legal && err != nil {
					t.Fatalf("%s should accept %s: %v", tc.card, tgt.typeLine, err)
				}
				if !tgt.legal {
					if err == nil {
						t.Fatalf("%s must reject %s", tc.card, tgt.typeLine)
					}
					return
				}
				effectstest.PassPriorityAroundTable(t, g)
				if g.Battlefield.Contains(id) {
					t.Errorf("%s survived %s", tgt.typeLine, tc.card)
				}
			})
		}
	}
}

// Despark's mana-value floor is the cost of its efficiency: a cheap
// threat and a land are both off limits.
func TestDesparkRespectsManaValueFloor(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, manaCost string
		legal                    bool
	}{
		{"Avacyn, Angel of Hope", "Creature — Angel", "{8}", true},
		{"Baneslayer Angel", "Creature — Angel", "{3}{W}{W}", true},
		{"Grizzly Bears", "Creature — Bear", "{1}{G}", false},
		{"Forest", "Basic Land — Forest", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			victim := g.Seats[1]
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: tc.name, TypeLine: tc.typeLine,
				ManaCost: tc.manaCost, Owner: victim.ID, Controller: victim.ID,
			})

			active := g.Seats[g.Turn.ActiveSeat]
			spellID := uuid.New()
			active.Hand.PushTop(game.Card{
				InstanceID: spellID, Name: "Despark", TypeLine: "Instant",
				OracleID: effectstest.DesparkOracle, Owner: active.ID, Controller: active.ID,
			})
			for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			err := g.CastSpell(active.ID, spellID, game.CastSpellParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
			})
			if tc.legal {
				if err != nil {
					t.Fatalf("Despark should accept mana value >= 4: %v", err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				if !g.Exile.Contains(id) {
					t.Error("target was not exiled")
				}
				return
			}
			if err == nil {
				t.Fatal("Despark must reject a permanent under mana value 4")
			}
		})
	}
}

// "You don't control" is enforced on both overload cards — your own
// board is never a legal target.
func TestOverloadCardsOnlyHitOpponents(t *testing.T) {
	for _, tc := range []struct {
		card, oracle, targetType string
	}{
		{"Vandalblast", effectstest.VandalblastOracle, "Artifact"},
		{"Cyclonic Rift", effectstest.CyclonicRiftOracle, "Creature — Bear"},
	} {
		t.Run(tc.card, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			caster := g.Seats[0]
			mine := effectstest.SeedPermanentFor(g, caster.ID, "Mine", tc.targetType)

			active := g.Seats[g.Turn.ActiveSeat]
			spellID := uuid.New()
			active.Hand.PushTop(game.Card{
				InstanceID: spellID, Name: tc.card, TypeLine: "Instant",
				OracleID: tc.oracle, Owner: active.ID, Controller: active.ID,
			})
			for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			if err := g.CastSpell(active.ID, spellID, game.CastSpellParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
			}); err == nil {
				t.Errorf("%s must not target a permanent you control", tc.card)
			}
		})
	}
}

func TestCyclonicRiftBouncesOpponentPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	id := effectstest.SeedPermanentFor(g, victim.ID, "Grizzly Bears", "Creature — Bear")
	handBefore := victim.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Cyclonic Rift", "Instant", effectstest.CyclonicRiftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(id) {
		t.Error("permanent was not bounced")
	}
	// The hand COUNT is a proxy: any other card arriving satisfies
	// it while this one went to a graveyard. Name the card.
	if !victim.Hand.Contains(id) {
		t.Error("the bounced permanent did not reach its owner's hand")
	}
	if victim.Hand.Size() != handBefore+1 {
		t.Errorf("owner's hand %d, want %d", victim.Hand.Size(), handBefore+1)
	}
}

func TestVandalblastDestroysOpponentArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	id := effectstest.SeedPermanentFor(g, victim.ID, "Sol Ring", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Vandalblast", "Sorcery", effectstest.VandalblastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(id) {
		t.Error("artifact survived Vandalblast")
	}
}

// --- Merciless Eviction -----------------------------------------

// Each mode sweeps exactly its own type and leaves the rest alone.
// Exile, not destroy — that's the reason to play it.
func TestMercilessEvictionSweepsChosenTypeOnly(t *testing.T) {
	for _, tc := range []struct {
		mode          int
		exiledTypes   []string
		survivedTypes []string
	}{
		{0, []string{"Artifact"}, []string{"Creature — Bear", "Enchantment"}},
		{1, []string{"Creature — Bear"}, []string{"Artifact", "Enchantment"}},
		{2, []string{"Enchantment"}, []string{"Artifact", "Creature — Bear"}},
	} {
		t.Run(tc.exiledTypes[0], func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			byType := map[string]uuid.UUID{}
			for _, tl := range []string{"Artifact", "Creature — Bear", "Enchantment"} {
				// One on each side of the table: the sweep is global.
				byType[tl] = effectstest.SeedPermanentFor(g, g.Seats[1].ID, "Victim", tl)
				effectstest.SeedPermanentFor(g, g.Seats[0].ID, "Mine", tl)
			}

			active := g.Seats[g.Turn.ActiveSeat]
			spellID := uuid.New()
			active.Hand.PushTop(game.Card{
				InstanceID: spellID, Name: "Merciless Eviction", TypeLine: "Sorcery",
				OracleID: effectstest.MercilessEvictionOracle, Owner: active.ID, Controller: active.ID,
			})
			for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			if err := g.CastSpell(active.ID, spellID, game.CastSpellParams{
				Modes: []int{tc.mode},
			}); err != nil {
				t.Fatalf("CastSpell mode %d: %v", tc.mode, err)
			}
			effectstest.PassPriorityAroundTable(t, g)

			for _, tl := range tc.exiledTypes {
				if !g.Exile.Contains(byType[tl]) {
					t.Errorf("%s was not exiled by mode %d", tl, tc.mode)
				}
			}
			for _, tl := range tc.survivedTypes {
				if !g.Battlefield.Contains(byType[tl]) {
					t.Errorf("%s should have survived mode %d", tl, tc.mode)
				}
			}
		})
	}
}
