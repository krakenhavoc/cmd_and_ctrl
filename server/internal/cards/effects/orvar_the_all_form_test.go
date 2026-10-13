package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// orvar_the_all_form_test.go — ADR 0153 PR 2 (#3168).

const orvarOracle = "276c4643-13ad-43bf-840b-08c025ebc065"

func (d dboTable) countByName(name string) int {
	n := 0
	for _, c := range d.g.BattlefieldCardsForEffect() {
		if c.Name == name {
			n++
		}
	}
	return n
}

// An opponent's discard: a token that's a copy of target permanent,
// chosen by the card's owner as the ability goes on the stack, and
// created under that owner's control.
func TestOrvarCopiesTargetPermanentWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	orvar := d.handCard(d.victim, "Orvar, the All-Form", "Legendary Creature — Shapeshifter", "{3}{U}", orvarOracle, 3, 3)
	filler := d.filler(d.victim)
	bear := b12Creature(d.g, d.caster.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, orvar, filler)
	passPriorityAroundTable(t, d.g)
	pickTargetRefs(t, d.g, d.victim.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, d.g)

	if n := d.countNamedUnder(d.victim, "Their Bear"); n != 1 {
		t.Errorf("the victim controls %d copies of the Bear, want 1", n)
	}
	if n := d.countNamedUnder(d.caster, "Their Bear"); n != 1 {
		t.Errorf("the original Bear's controller has %d, want just the original", n)
	}
	if !d.victim.Graveyard.Contains(orvar) {
		t.Errorf("Orvar is in %q, want the graveyard", d.zoneOf(orvar))
	}
}

func TestOrvarDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	orvar := d.handCard(d.caster, "Orvar, the All-Form", "Legendary Creature — Shapeshifter", "{3}{U}", orvarOracle, 3, 3)
	filler := d.filler(d.caster)
	b12Creature(d.g, d.victim.ID, "Their Bear", "Creature — Bear", 2, 2)

	d.iDiscard(t, orvar, filler)

	if latestPickTarget(d.g, d.caster.ID) != nil {
		t.Error("your own discard put the trigger on the stack")
	}
	if n := d.countByName("Their Bear"); n != 1 {
		t.Errorf("Bears = %d, want 1", n)
	}
}

// "Whenever you cast an instant or sorcery spell, if it targets one or
// more other permanents you control, create a token that's a copy of one
// of those permanents": a Bear is copied; a spell aimed at a player or at
// an opponent's permanent copies nothing.
func TestOrvarCopiesAPermanentYourSpellTargets(t *testing.T) {
	d := newDBOTable(t)
	apaPush(d.g, d.caster.ID, d.caster.ID, apaCreature("Orvar, the All-Form", orvarOracle, 3, 3))
	mine := b12Creature(d.g, d.caster.ID, "My Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, d.g, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, d.g)

	if n := d.countNamedUnder(d.caster, "My Bear"); n != 2 {
		t.Errorf("My Bears = %d, want the original and a token copy", n)
	}
}

func TestOrvarIgnoresSpellsThatTargetNothingOfYours(t *testing.T) {
	d := newDBOTable(t)
	orvar := apaPush(d.g, d.caster.ID, d.caster.ID, apaCreature("Orvar, the All-Form", orvarOracle, 3, 3))
	theirs := b12Creature(d.g, d.victim.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, d.g, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, d.g)
	castCatalogSpell(t, d.g, "Shock", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	castCatalogSpell(t, d.g, "Self Pump", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: orvar}})
	passPriorityAroundTable(t, d.g)

	if n := d.countByName("Their Bear"); n != 1 {
		t.Errorf("Their Bears = %d, want 1: an opponent's permanent is not copied", n)
	}
	if n := d.countByName("Orvar, the All-Form"); n != 1 {
		t.Errorf("Orvars = %d, want 1: \"other\" excludes Orvar itself", n)
	}
}

// An Aura spell (not an instant or sorcery) does not trigger it.
func TestOrvarIgnoresSpellsThatAreNotInstantsOrSorceries(t *testing.T) {
	d := newDBOTable(t)
	apaPush(d.g, d.caster.ID, d.caster.ID, apaCreature("Orvar, the All-Form", orvarOracle, 3, 3))
	mine := b12Creature(d.g, d.caster.ID, "My Bear", "Creature — Bear", 2, 2)

	orvarPlainSpell(t, d, "Aura Creature", "Enchantment — Aura", mine)

	if n := d.countNamedUnder(d.caster, "My Bear"); n != 1 {
		t.Errorf("My Bears = %d, want 1: only instants and sorceries trigger it", n)
	}
}

func orvarPlainSpell(t *testing.T, d dboTable, name, typeLine string, target uuid.UUID) {
	t.Helper()
	_ = castCatalogSpellErr(t, d.g, name, typeLine, "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, d.g)
}
