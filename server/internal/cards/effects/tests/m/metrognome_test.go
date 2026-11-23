package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An opponent's discard: four 1/1 colorless Gnome artifact creature
// tokens for the card's owner.
func TestMetrognomeMakesFourGnomesWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	m := d.HandCard(d.Victim, "Metrognome", "Artifact", "{4}", effectstest.MetrognomeOracle, 0, 0)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, m, filler)

	if n := d.Gnomes(d.Victim); n != 4 {
		t.Fatalf("Gnomes = %d, want 4", n)
	}
	for _, c := range d.G.BattlefieldCardsForEffect() {
		if c.Name == "Gnome" && (c.Power != 1 || c.Toughness != 1 || !c.IsCreature() || !c.IsArtifact()) {
			t.Errorf("a Gnome is %d/%d creature=%v artifact=%v", c.Power, c.Toughness, c.IsCreature(), c.IsArtifact())
		}
	}
	if !d.Victim.Graveyard.Contains(m) {
		t.Errorf("Metrognome is in %q, want the graveyard", d.ZoneOf(m))
	}
}

func TestMetrognomeMakesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	m := d.HandCard(d.Caster, "Metrognome", "Artifact", "{4}", effectstest.MetrognomeOracle, 0, 0)
	filler := d.Filler(d.Caster)

	d.IDiscard(t, m, filler)

	if n := d.Gnomes(d.Caster); n != 0 {
		t.Errorf("Gnomes = %d, want 0", n)
	}
}

// {4}, {T}: one more Gnome.
func TestMetrognomeActivatedAbilityMakesAGnome(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	m := effectstest.ApaPush(d.G, d.Caster.ID, d.Caster.ID, game.Card{Name: "Metrognome", TypeLine: "Artifact", OracleID: effectstest.MetrognomeOracle})

	effectstest.ApaMana(d.Caster, "R", "R", "R", "R")
	effectstest.ApaActivate(t, d.G, d.Caster, m, 0, game.ActivateAbilityParams{})

	if n := d.Gnomes(d.Caster); n != 1 {
		t.Errorf("Gnomes = %d, want 1", n)
	}
}

// The trigger declares what it pays for the bot: four tokens.
func TestMetrognomeDeclaresItsDiscardPayoff(t *testing.T) {
	for _, r := range game.AbilityRowsOf(game.Card{OracleID: effectstest.MetrognomeOracle}) {
		if r.Purpose.OnOpponentDiscard != nil && r.Purpose.OnOpponentDiscard.Tokens == 4 {
			return
		}
	}
	t.Error("no triggered row declares OnOpponentDiscard with four tokens")
}
