package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// metrognome_test.go — ADR 0153 PR 2 (#3168).

const metrognomeOracle = "57fe5c94-ee93-45c6-bf7e-89d177d0fe70"

func (d dboTable) gnomes(p *game.Player) int { return d.countNamedUnder(p, "Gnome") }

// An opponent's discard: four 1/1 colorless Gnome artifact creature
// tokens for the card's owner.
func TestMetrognomeMakesFourGnomesWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	m := d.handCard(d.victim, "Metrognome", "Artifact", "{4}", metrognomeOracle, 0, 0)
	filler := d.filler(d.victim)

	d.theyDiscard(t, m, filler)

	if n := d.gnomes(d.victim); n != 4 {
		t.Fatalf("Gnomes = %d, want 4", n)
	}
	for _, c := range d.g.BattlefieldCardsForEffect() {
		if c.Name == "Gnome" && (c.Power != 1 || c.Toughness != 1 || !c.IsCreature() || !c.IsArtifact()) {
			t.Errorf("a Gnome is %d/%d creature=%v artifact=%v", c.Power, c.Toughness, c.IsCreature(), c.IsArtifact())
		}
	}
	if !d.victim.Graveyard.Contains(m) {
		t.Errorf("Metrognome is in %q, want the graveyard", d.zoneOf(m))
	}
}

func TestMetrognomeMakesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	m := d.handCard(d.caster, "Metrognome", "Artifact", "{4}", metrognomeOracle, 0, 0)
	filler := d.filler(d.caster)

	d.iDiscard(t, m, filler)

	if n := d.gnomes(d.caster); n != 0 {
		t.Errorf("Gnomes = %d, want 0", n)
	}
}

// {4}, {T}: one more Gnome.
func TestMetrognomeActivatedAbilityMakesAGnome(t *testing.T) {
	d := newDBOTable(t)
	m := apaPush(d.g, d.caster.ID, d.caster.ID, game.Card{Name: "Metrognome", TypeLine: "Artifact", OracleID: metrognomeOracle})

	apaMana(d.caster, "R", "R", "R", "R")
	apaActivate(t, d.g, d.caster, m, 0, game.ActivateAbilityParams{})

	if n := d.gnomes(d.caster); n != 1 {
		t.Errorf("Gnomes = %d, want 1", n)
	}
}

// The trigger declares what it pays for the bot: four tokens.
func TestMetrognomeDeclaresItsDiscardPayoff(t *testing.T) {
	for _, r := range game.AbilityRowsOf(game.Card{OracleID: metrognomeOracle}) {
		if r.Purpose.OnOpponentDiscard != nil && r.Purpose.OnOpponentDiscard.Tokens == 4 {
			return
		}
	}
	t.Error("no triggered row declares OnOpponentDiscard with four tokens")
}
