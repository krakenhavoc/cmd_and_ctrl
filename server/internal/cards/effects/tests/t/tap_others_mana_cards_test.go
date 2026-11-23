package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	springleafDrumOracle    = "bdeb440d-a714-40e8-9038-f651e6ae45fb"
	jasperaSentinelOracle   = "e2f98813-4f39-4135-b05b-407e9eb797f6"
	heritageDruidOracle     = "0b7f9c71-6f3c-4056-9bed-04b9f2296f3b"
	relicOfLegendsOracle    = "4d9554de-c005-41ab-a941-39328ebff9de"
	holdoutSettlementOracle = "e6b77545-de5c-4f4a-b7ea-83498fb33ba8"
)

func TestJasperaSentinelPaysItsTwoTapComponentsWithDifferentCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sentinel := effectstest.SeedPermanentWithOracle(g, me.ID, "Jaspera Sentinel", "Creature — Elf Rogue", jasperaSentinelOracle)

	if err := g.ActivateManaAbility(me.ID, sentinel, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{sentinel}}); err == nil {
		t.Fatal("Jaspera Sentinel paid both {T} and its creature-tap component with itself")
	}
	effectstest.AssertTapped(t, g, false, sentinel)

	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Elf", 1, 1)
	if err := g.ActivateManaAbility(me.ID, sentinel, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{friend}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	effectstest.AssertTapped(t, g, true, sentinel, friend)
	if effectstest.ManaPickFor(g, me.ID) == nil {
		t.Fatal("the any-color ability did not ask for a color")
	}
}

func TestHeritageDruidTapsThreeElvesIncludingItselfForThreeGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.SeedPermanentWithOracle(g, me.ID, "Heritage Druid", "Creature — Elf Druid", heritageDruidOracle)
	one := effectstest.B12Creature(g, me.ID, "Elf One", "Creature — Elf", 1, 1)
	two := effectstest.B12Creature(g, me.ID, "Elf Two", "Creature — Elf", 1, 1)

	if err := g.ActivateManaAbility(me.ID, druid, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{druid, one, two}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	effectstest.AssertTapped(t, g, true, druid, one, two)
	if me.ManaPool.Total() != 3 {
		t.Fatalf("mana pool = %v, want three green mana", me.ManaPool)
	}
	for _, mana := range me.ManaPool.Units() {
		if mana.Color != "G" {
			t.Errorf("mana pool = %v, want only green mana", me.ManaPool)
			break
		}
	}
}

func TestRelicOfLegendsTapsOnlyALegendaryCreatureAndNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.SeedPermanentWithOracle(g, me.ID, "Relic of Legends", "Artifact", relicOfLegendsOracle)
	ordinary := effectstest.B12Creature(g, me.ID, "Ordinary Bear", "Creature — Bear", 2, 2)
	legend := effectstest.B12Creature(g, me.ID, "Captain", "Legendary Creature — Human", 2, 2)

	if err := g.ActivateManaAbility(me.ID, relic, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{ordinary}}); err == nil {
		t.Fatal("Relic of Legends accepted a nonlegendary creature")
	}
	effectstest.AssertTapped(t, g, false, relic, ordinary, legend)

	if err := g.ActivateManaAbility(me.ID, relic, 1, game.ManaAbilityParams{TapIDs: []uuid.UUID{legend}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	effectstest.AssertTapped(t, g, false, relic)
	effectstest.AssertTapped(t, g, true, legend)
}

func TestFixedCountTapOthersManaCardsAreFull(t *testing.T) {
	for oracleID, name := range map[string]string{
		springleafDrumOracle:    "Springleaf Drum",
		jasperaSentinelOracle:   "Jaspera Sentinel",
		heritageDruidOracle:     "Heritage Druid",
		relicOfLegendsOracle:    "Relic of Legends",
		holdoutSettlementOracle: "Holdout Settlement",
	} {
		spec, ok := effects.Lookup(oracleID)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s registration = %q / %q, want full", name, spec.Name, spec.Completeness)
		}
	}
}
