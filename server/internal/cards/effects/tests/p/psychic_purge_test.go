package testsp

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPsychicPurgeCastDealsOne(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	life := d.Victim.Life

	effectstest.CastCatalogSpell(t, d.G, "Psychic Purge", "Sorcery", effectstest.PsychicPurgeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)

	if d.Victim.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, d.Victim.Life)
	}
}

// "That player" is the controller of the spell that made you discard:
// the caster of the Mind Rot loses 5 life, and the card's owner does not.
func TestPsychicPurgeMakesThePlayerWhoCausedTheDiscardLoseFive(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	purge := d.HandCard(d.Victim, "Psychic Purge", "Sorcery", "{U}", effectstest.PsychicPurgeOracle, 0, 0)
	filler := d.Filler(d.Victim)
	casterLife, victimLife := d.Caster.Life, d.Victim.Life

	d.TheyDiscard(t, purge, filler)

	if d.Caster.Life != casterLife-5 {
		t.Errorf("the Mind Rot's controller has %d life, want %d", d.Caster.Life, casterLife-5)
	}
	if d.Victim.Life != victimLife {
		t.Errorf("the card's owner has %d life, want %d", d.Victim.Life, victimLife)
	}
}

func TestPsychicPurgeDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	purge := d.HandCard(d.Caster, "Psychic Purge", "Sorcery", "{U}", effectstest.PsychicPurgeOracle, 0, 0)
	filler := d.Filler(d.Caster)
	life := d.Caster.Life

	d.IDiscard(t, purge, filler)

	if d.Caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.Caster.Life)
	}
}
