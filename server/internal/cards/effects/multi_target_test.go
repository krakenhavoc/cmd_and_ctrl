package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multi_target_test.go — S20 sub-PR 5: clauses with more than one
// target slot.

const (
	arcTrailOracle          = "f1c26b25-371e-4fbf-a43d-7fd59a364d3a"
	ashesToAshesOracle      = "944a52d9-bb14-43c5-8d05-2afaf023dc9f"
	sylvanReclamationOracle = "aeec8e85-6571-4da6-8a48-f5d3985ca10b"
)

func cardRefs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func TestMultiTargetSpecsAreWired(t *testing.T) {
	cases := map[string][2]int{arcTrailOracle: {2, 2}, ashesToAshesOracle: {2, 2}, sylvanReclamationOracle: {0, 2}}
	for oracle, want := range cases {
		spec := game.TargetSpecFor(oracle)
		if spec == nil || spec.Min != want[0] || spec.Max != want[1] {
			t.Errorf("%s: spec %+v, want count %v", oracle, spec, want)
		}
	}
}

// --- Arc Trail ---------------------------------------------------

func TestArcTrailIsPositional(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := pushCostedPermanentForTest(g, opp.ID, "Big", "Creature — Beast", "{2}{G}")
	before := opp.Life
	castCatalogSpell(t, g, "Arc Trail", "Sorcery", arcTrailOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("first slot: life %d -> %d, want -2", before, opp.Life)
	}
	// 2/2 took 1: survives.
	if !g.Battlefield.Contains(big) {
		t.Errorf("second slot took 2 instead of 1")
	}
}

func TestArcTrailRejectsSameTargetTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Arc Trail", TypeLine: "Sorcery",
		OracleID: arcTrailOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: opp.ID}}})
	if err != game.ErrInvalidParam {
		t.Fatalf("same player twice: %v, want ErrInvalidParam", err)
	}
	err = g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if err != game.ErrInvalidParam {
		t.Fatalf("one target: %v, want ErrInvalidParam", err)
	}
}

// One slot gone at resolution: the other still lands.
func TestArcTrailPartialTargetStillResolves(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushCostedPermanentForTest(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	before := opp.Life
	id := castCatalogSpell(t, g, "Arc Trail", "Sorcery", arcTrailOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID}})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
	passPriorityAroundTable(t, g)
	if g.Seats[0].Graveyard.Contains(id) == false {
		t.Errorf("Arc Trail should have resolved to the graveyard")
	}
	if opp.Life != before-1 {
		t.Errorf("second slot with first gone: life %d -> %d, want -1", before, opp.Life)
	}
}

// --- Ashes to Ashes ----------------------------------------------

func TestAshesToAshesExilesTwoAndBurnsCaster(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCostedPermanentForTest(g, opp.ID, "A", "Creature — Bear", "{1}{G}")
	b := pushCostedPermanentForTest(g, me.ID, "B", "Creature — Bear", "{1}{G}")
	c := pushCostedPermanentForTest(g, opp.ID, "C", "Creature — Bear", "{1}{G}")
	before := me.Life
	castCatalogSpell(t, g, "Ashes to Ashes", "Sorcery", ashesToAshesOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Errorf("both targets should be exiled")
	}
	if !g.Battlefield.Contains(c) {
		t.Errorf("untargeted creature must stay")
	}
	if me.Life != before-5 {
		t.Errorf("caster life %d -> %d, want -5", before, me.Life)
	}
}

func TestAshesToAshesRejectsArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := pushCostedPermanentForTest(g, opp.ID, "Golem", "Artifact Creature — Golem", "{3}")
	bear := pushCostedPermanentForTest(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ashes to Ashes", TypeLine: "Sorcery",
		OracleID: ashesToAshesOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(bear, golem)}); err != game.ErrIllegalTarget {
		t.Fatalf("artifact creature: %v, want ErrIllegalTarget", err)
	}
	lt := g.LegalTargetsFor(game.SourceChooser(me.ID), ashesToAshesOracle)
	if lt == nil || len(lt.Cards) != 1 || lt.Cards[0] != bear {
		t.Errorf("legal set = %+v, want just the bear", lt)
	}
}

// One of the two exiled in response: the other is exiled and the
// caster still takes 5.
func TestAshesToAshesPartial(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushCostedPermanentForTest(g, opp.ID, "A", "Creature — Bear", "{1}{G}")
	b := pushCostedPermanentForTest(g, opp.ID, "B", "Creature — Bear", "{1}{G}")
	before := me.Life
	castCatalogSpell(t, g, "Ashes to Ashes", "Sorcery", ashesToAshesOracle, cardRefs(a, b))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(b) {
		t.Errorf("surviving target should be exiled")
	}
	if !opp.Graveyard.Contains(a) {
		t.Errorf("destroyed target should stay in the graveyard, not be exiled")
	}
	if me.Life != before-5 {
		t.Errorf("caster life %d -> %d, want -5 even with one target gone", before, me.Life)
	}
}

// --- Sylvan Reclamation ------------------------------------------

func TestSylvanReclamationUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushCostedPermanentForTest(g, opp.ID, "Rock", "Artifact", "{1}")
	aura := pushCostedPermanentForTest(g, opp.ID, "Aura", "Enchantment", "{G}")
	bear := pushCostedPermanentForTest(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	// Zero targets is a legal cast.
	none := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: none, Name: "Sylvan Reclamation", TypeLine: "Instant",
		OracleID: sylvanReclamationOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, none, game.CastSpellParams{}); err != nil {
		t.Fatalf("up to two with none: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(aura) {
		t.Fatalf("no-target cast must exile nothing")
	}
	// Three is too many; a creature is illegal.
	three := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: three, Name: "Sylvan Reclamation", TypeLine: "Instant",
		OracleID: sylvanReclamationOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, three, game.CastSpellParams{Targets: cardRefs(rock, aura, bear)}); err != game.ErrIllegalTarget {
		t.Fatalf("creature among targets: %v, want ErrIllegalTarget", err)
	}
	// Two: both exiled.
	if err := g.CastSpell(me.ID, three, game.CastSpellParams{Targets: cardRefs(rock, aura)}); err != nil {
		t.Fatalf("two targets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) || !g.Exile.Contains(aura) {
		t.Errorf("both should be exiled")
	}
	if !g.Battlefield.Contains(bear) {
		t.Errorf("creature untouched")
	}
}
