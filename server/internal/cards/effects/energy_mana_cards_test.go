package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// energy_mana_cards_test.go — ADR 0129 PR 2's cards: the mana abilities
// that pay energy (Aether Hub, Servant of the Conduit, Solar
// Transformer, Conversion Apparatus) and the auto-tapper's energy tier.

const (
	aetherHubOracle         = "61c89b11-65c9-4fda-bbcd-d84de25df801"
	servantOfConduitOracle  = "df1a8846-b4da-4e9b-8cf3-ee687e4c606b"
	solarTransformerOracle  = "cb377c27-39a6-42b4-8c89-80a9a2350e81"
	conversionApparatOracle = "931d3dcb-4bbb-4f1a-95e2-7e315faf4158"
)

// Each card is Full and declares the energy it costs on the right row.
func TestEnergyManaCardsDeclareTheirCosts(t *testing.T) {
	for oracle, want := range map[string]int{
		aetherHubOracle: 1, servantOfConduitOracle: 1, solarTransformerOracle: 1, conversionApparatOracle: 3,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		last := spec.ManaAbilities[len(spec.ManaAbilities)-1]
		if last.Cost.Energy != want || !last.Cost.Tap {
			t.Errorf("%s: last mana row cost %+v, want {T} and %d energy", spec.Name, last.Cost, want)
		}
	}
}

// Aether Hub's enters trigger gives one energy, and the auto-tapped cast
// of a blue spell spends it through the Hub's coloured half, which the
// move list offers only while the energy is there.
func TestAetherHubPaysAColouredPipWithEnergy(t *testing.T) {
	g, me, _ := spendTable(t)
	hub := pushCatalogPermanent(g, me.ID, "Aether Hub", "Land", aetherHubOracle, false)
	spell := handSpell(me, "Blue Spell", "Instant", "{U}")

	setEnergy(t, g, me, 0)
	if castMove(g, me.ID, spell) {
		t.Fatal("a {U} spell is offered with no energy for the Hub")
	}
	setEnergy(t, g, me, 1)
	if !castMove(g, me.ID, spell) {
		t.Fatal("a {U} spell is not offered with one energy for the Hub")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped cast off Aether Hub: %v", err)
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0 after the Hub paid", energyOf(me))
	}
	if !isTapped(g, hub) {
		t.Error("Aether Hub did not tap")
	}
}

// The {C} half pays a generic pip without touching the energy.
func TestAetherHubPaysGenericWithoutEnergy(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Aether Hub", "Land", aetherHubOracle, false)
	spell := handSpell(me, "Grey Spell", "Artifact", "{1}")
	setEnergy(t, g, me, 1)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1: the {C} half pays a generic pip", energyOf(me))
	}
}

// Energy before life: a Hub and a Mana Confluence, a {U} spell — the Hub
// pays and no life is lost.
func TestAetherHubPaysBeforeManaConfluence(t *testing.T) {
	g, me, _ := spendTable(t)
	confluence := pushCatalogPermanent(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle, false)
	pushCatalogPermanent(g, me.ID, "Aether Hub", "Land", aetherHubOracle, false)
	spell := handSpell(me, "Blue Spell", "Instant", "{U}")
	setEnergy(t, g, me, 1)
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != life || isTapped(g, confluence) {
		t.Errorf("life %d → %d, Confluence tapped %v: the Hub's energy should have paid", life, me.Life, isTapped(g, confluence))
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
}

// Conversion Apparatus: "{3}, {T}: You get {E}{E}{E}" is an ordinary
// activation, and its mana ability pays three energy for three mana.
func TestConversionApparatusMakesAndSpendsEnergy(t *testing.T) {
	g, me, _ := spendTable(t)
	app := pushCatalogPermanent(g, me.ID, "Conversion Apparatus", "Artifact", conversionApparatOracle, false)

	setEnergy(t, g, me, 2)
	err := g.ActivateManaAbility(me.ID, app, 1, game.ManaAbilityParams{Colors: []string{"W", "U", "B"}})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("two energy: err = %v, want ErrInsufficientEnergy", err)
	}
	if isTapped(g, app) {
		t.Fatal("a refused activation tapped the Apparatus")
	}
	setEnergy(t, g, me, 3)
	if err := g.ActivateManaAbility(me.ID, app, 1, game.ManaAbilityParams{Colors: []string{"W", "U", "B"}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if energyOf(me) != 0 || len(me.ManaPool) != 3 {
		t.Errorf("energy %d, pool %v; want 0 and three mana", energyOf(me), me.ManaPool)
	}
}

func TestConversionApparatusGetsThreeEnergy(t *testing.T) {
	g, me, _ := spendTable(t)
	app := pushCatalogPermanent(g, me.ID, "Conversion Apparatus", "Artifact", conversionApparatOracle, false)
	apaMana(me, "C", "C", "C")
	setEnergy(t, g, me, 0)
	p7Activate(t, g, me, app, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 3 {
		t.Errorf("energy = %d, want 3", energyOf(me))
	}
}

// The enumerator offers the energy mana ability only with the energy,
// and names it on the move's cost.
func TestEnergyManaAbilityMoveNamesTheEnergy(t *testing.T) {
	g, me, _ := spendTable(t)
	hub := pushCatalogPermanent(g, me.ID, "Aether Hub", "Land", aetherHubOracle, false)
	energyMoves := func() []legal.Move {
		var out []legal.Move
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == "activate_mana_ability" && m.Source == hub && m.Cost != nil && m.Cost.Energy > 0 {
				out = append(out, m)
			}
		}
		return out
	}
	setEnergy(t, g, me, 0)
	if ms := energyMoves(); len(ms) != 0 {
		t.Errorf("with no energy the Hub's coloured half is offered: %d moves", len(ms))
	}
	setEnergy(t, g, me, 1)
	ms := energyMoves()
	if len(ms) == 0 {
		t.Fatal("with one energy the Hub's coloured half is not offered")
	}
	if ms[0].Cost.Energy != 1 {
		t.Errorf("move cost energy = %d, want 1", ms[0].Cost.Energy)
	}
}
