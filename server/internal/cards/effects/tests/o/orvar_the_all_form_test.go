package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An opponent's discard: a token that's a copy of target permanent,
// chosen by the card's owner as the ability goes on the stack, and
// created under that owner's control.
func TestOrvarCopiesTargetPermanentWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	orvar := d.HandCard(d.Victim, "Orvar, the All-Form", "Legendary Creature — Shapeshifter", "{3}{U}", effectstest.OrvarOracle, 3, 3)
	filler := d.Filler(d.Victim)
	bear := effectstest.B12Creature(d.G, d.Caster.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, orvar, filler)
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.PickTargetRefs(t, d.G, d.Victim.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, d.G)

	if n := d.CountNamedUnder(d.Victim, "Their Bear"); n != 1 {
		t.Errorf("the victim controls %d copies of the Bear, want 1", n)
	}
	if n := d.CountNamedUnder(d.Caster, "Their Bear"); n != 1 {
		t.Errorf("the original Bear's controller has %d, want just the original", n)
	}
	if !d.Victim.Graveyard.Contains(orvar) {
		t.Errorf("Orvar is in %q, want the graveyard", d.ZoneOf(orvar))
	}
}

func TestOrvarDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	orvar := d.HandCard(d.Caster, "Orvar, the All-Form", "Legendary Creature — Shapeshifter", "{3}{U}", effectstest.OrvarOracle, 3, 3)
	filler := d.Filler(d.Caster)
	effectstest.B12Creature(d.G, d.Victim.ID, "Their Bear", "Creature — Bear", 2, 2)

	d.IDiscard(t, orvar, filler)

	if effectstest.LatestPickTarget(d.G, d.Caster.ID) != nil {
		t.Error("your own discard put the trigger on the stack")
	}
	if n := d.CountByName("Their Bear"); n != 1 {
		t.Errorf("Bears = %d, want 1", n)
	}
}

// "Whenever you cast an instant or sorcery spell, if it targets one or
// more other permanents you control, create a token that's a copy of one
// of those permanents": a Bear is copied; a spell aimed at a player or at
// an opponent's permanent copies nothing.
func TestOrvarCopiesAPermanentYourSpellTargets(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.ApaPush(d.G, d.Caster.ID, d.Caster.ID, effectstest.ApaCreature("Orvar, the All-Form", effectstest.OrvarOracle, 3, 3))
	mine := effectstest.B12Creature(d.G, d.Caster.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, d.G, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, d.G)

	if n := d.CountNamedUnder(d.Caster, "My Bear"); n != 2 {
		t.Errorf("My Bears = %d, want the original and a token copy", n)
	}
}

func TestOrvarIgnoresSpellsThatTargetNothingOfYours(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	orvar := effectstest.ApaPush(d.G, d.Caster.ID, d.Caster.ID, effectstest.ApaCreature("Orvar, the All-Form", effectstest.OrvarOracle, 3, 3))
	theirs := effectstest.B12Creature(d.G, d.Victim.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, d.G, "Giant Growth", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.CastCatalogSpell(t, d.G, "Shock", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.CastCatalogSpell(t, d.G, "Self Pump", "Instant", "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: orvar}})
	effectstest.PassPriorityAroundTable(t, d.G)

	if n := d.CountByName("Their Bear"); n != 1 {
		t.Errorf("Their Bears = %d, want 1: an opponent's permanent is not copied", n)
	}
	if n := d.CountByName("Orvar, the All-Form"); n != 1 {
		t.Errorf("Orvars = %d, want 1: \"other\" excludes Orvar itself", n)
	}
}

// An Aura spell (not an instant or sorcery) does not trigger it.
func TestOrvarIgnoresSpellsThatAreNotInstantsOrSorceries(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	effectstest.ApaPush(d.G, d.Caster.ID, d.Caster.ID, effectstest.ApaCreature("Orvar, the All-Form", effectstest.OrvarOracle, 3, 3))
	mine := effectstest.B12Creature(d.G, d.Caster.ID, "My Bear", "Creature — Bear", 2, 2)

	orvarPlainSpell(t, d, "Aura Creature", "Enchantment — Aura", mine)

	if n := d.CountNamedUnder(d.Caster, "My Bear"); n != 1 {
		t.Errorf("My Bears = %d, want 1: only instants and sorceries trigger it", n)
	}
}

func orvarPlainSpell(t *testing.T, d effectstest.DboTable, name, typeLine string, target uuid.UUID) {
	t.Helper()
	_ = effectstest.CastCatalogSpellErr(t, d.G, name, typeLine, "not-in-catalog",
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, d.G)
}
