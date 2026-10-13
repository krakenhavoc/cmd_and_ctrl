package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// psychic_purge_test.go — ADR 0153 PR 2 (#3168).

const psychicPurgeOracle = "6af596d2-f075-4f55-b088-a5237fcdaa51"

func TestPsychicPurgeCastDealsOne(t *testing.T) {
	d := newDBOTable(t)
	life := d.victim.Life

	castCatalogSpell(t, d.g, "Psychic Purge", "Sorcery", psychicPurgeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)

	if d.victim.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, d.victim.Life)
	}
}

// "That player" is the controller of the spell that made you discard:
// the caster of the Mind Rot loses 5 life, and the card's owner does not.
func TestPsychicPurgeMakesThePlayerWhoCausedTheDiscardLoseFive(t *testing.T) {
	d := newDBOTable(t)
	purge := d.handCard(d.victim, "Psychic Purge", "Sorcery", "{U}", psychicPurgeOracle, 0, 0)
	filler := d.filler(d.victim)
	casterLife, victimLife := d.caster.Life, d.victim.Life

	d.theyDiscard(t, purge, filler)

	if d.caster.Life != casterLife-5 {
		t.Errorf("the Mind Rot's controller has %d life, want %d", d.caster.Life, casterLife-5)
	}
	if d.victim.Life != victimLife {
		t.Errorf("the card's owner has %d life, want %d", d.victim.Life, victimLife)
	}
}

func TestPsychicPurgeDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	purge := d.handCard(d.caster, "Psychic Purge", "Sorcery", "{U}", psychicPurgeOracle, 0, 0)
	filler := d.filler(d.caster)
	life := d.caster.Life

	d.iDiscard(t, purge, filler)

	if d.caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.caster.Life)
	}
}
