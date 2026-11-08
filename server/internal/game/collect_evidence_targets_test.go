package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// collect_evidence_targets_test.go — #5240: "collect evidence X, where X
// is the total mana value of the permanents this spell targets" (Urgent
// Necropsy) as a spell's mandatory additional cost. The targets are chosen
// at CR 601.2c and the total cost determined at CR 601.2f, so CastSpell
// fixes X from the announced targets and checks the cards on exile_ids
// against it (CR 701.59a, 118.3).

const ceTargetsOracle = "test-5240-collect-evidence-targets"

// ceTargetsTable is a main phase on seat 0's turn with a test spell in
// hand whose mandatory additional cost is the target-sized clause, plus
// whatever `extra` adds to it. The spell declares no target clause, so the
// test's targets are taken as sent.
func ceTargetsTable(t *testing.T, extra func(*AdditionalCost)) (*Game, *Player, uuid.UUID) {
	t.Helper()
	withCatalogAdditionalCost(t, func(o string) *AdditionalCost {
		if o != ceTargetsOracle {
			return nil
		}
		ac := &AdditionalCost{
			ExileCards: CollectEvidenceFromTargetsCost("cards from your graveyard with total mana value X or greater"),
			Label:      "Collect evidence X",
		}
		if extra != nil {
			extra(ac)
		}
		return ac
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	spell := NewCard("Test Necropsy", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = ceTargetsOracle
	me.Hand.PushTop(spell)
	return g, me, spell.InstanceID
}

// ceCard pushes a card with mana cost `cost` into z, owned and controlled
// by `owner`.
func ceCard(z *Zone, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	c := NewCard(name, owner)
	c.TypeLine = typeLine
	c.ManaCost = cost
	c.Controller = owner
	z.PushTop(c)
	return c.InstanceID
}

func ceRefs(ids ...uuid.UUID) []TargetRef {
	out := make([]TargetRef, len(ids))
	for i, id := range ids {
		out[i] = TargetRef{Kind: TargetCard, ID: id}
	}
	return out
}

// X is the total over the distinct permanents among the targets: a
// permanent named for two "target" words counts once (CR 115.3), and a
// player, a card in a graveyard and a face-down permanent's cost add
// nothing (CR 708.2a, 202.3a).
func TestCollectEvidenceTargetsFloorCountsEachPermanentOnce(t *testing.T) {
	g, me, _ := ceTargetsTable(t, nil)
	opp := g.Seats[1]
	three := ceCard(g.Battlefield, opp.ID, "Golem", "Artifact Creature — Golem", "{3}")
	two := ceCard(g.Battlefield, opp.ID, "Charm", "Enchantment", "{1}{W}")
	faceDown := ceCard(g.Battlefield, opp.ID, "Hidden", "Creature — Bear", "{5}{G}")
	inGrave := ceCard(me.Graveyard, me.ID, "Dead", "Creature — Bear", "{4}")
	g.WithWriteLock(func() {
		g.findBattlefieldCardLocked(faceDown).SetFaceDown(FaceDownMorphed)
	})
	refs := append(ceRefs(three, three, two, faceDown, inGrave), TargetRef{Kind: TargetPlayer, ID: opp.ID})
	var got int
	var mvs map[uuid.UUID]int
	g.WithWriteLock(func() {
		got = g.CollectEvidenceTargetsFloorForEffect(refs)
		mvs = g.PermanentManaValuesForEffect()
	})
	if got != 5 {
		t.Errorf("floor = %d, want 5 (the Golem once, the Charm, nothing for the rest)", got)
	}
	if mvs[three] != 3 || mvs[two] != 2 || mvs[faceDown] != 0 {
		t.Errorf("per-permanent mana values %v, want Golem 3, Charm 2, face-down 0", mvs)
	}
	if _, ok := mvs[inGrave]; ok {
		t.Error("a graveyard card is among the permanents' mana values")
	}
	g.WithWriteLock(func() {
		if n := g.CollectEvidenceTargetsFloorForEffect(nil); n != 0 {
			t.Errorf("no targets: floor %d, want 0", n)
		}
	})
}

// The announce gate: a pick below the X the targets fix is refused with
// nothing paid; one that reaches it is exiled with the spell on the stack
// and is one collection of evidence. With no targets X is 0 and the spell
// is cast exiling nothing, which is still collecting evidence (CR 701.59a).
func TestCollectEvidenceTargetsFloorIsCheckedAtAnnounce(t *testing.T) {
	g, me, spell := ceTargetsTable(t, nil)
	opp := g.Seats[1]
	golem := ceCard(g.Battlefield, opp.ID, "Golem", "Artifact Creature — Golem", "{3}")
	one := ceCard(me.Graveyard, me.ID, "One", "Sorcery", "{R}")
	two := ceCard(me.Graveyard, me.ID, "Two", "Sorcery", "{1}{R}")

	for name, ids := range map[string][]uuid.UUID{
		"nothing":                  nil,
		"cards worth 2 of 3":       {two},
		"the same card twice":      {two, two},
		"a card that is not there": {two, uuid.New()},
	} {
		if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: ceRefs(golem), ExileIDs: ids}); err == nil {
			t.Fatalf("%s paid collect evidence 3", name)
		}
	}
	if !me.Hand.Contains(spell) || !me.Graveyard.Contains(one) || !me.Graveyard.Contains(two) {
		t.Fatal("a refused cast paid something")
	}

	if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: ceRefs(golem), ExileIDs: []uuid.UUID{two, one}}); err != nil {
		t.Fatalf("cards worth 3 for a target worth 3: %v", err)
	}
	if !g.Stack.Contains(spell) || !g.Exile.Contains(one) || !g.Exile.Contains(two) {
		t.Fatal("the evidence was not exiled with the spell on the stack")
	}
	if item := g.StackMeta[spell]; item == nil || !slices.Equal(item.Paid.Exiled, []uuid.UUID{two, one}) {
		t.Errorf("the paid record does not hold the exiled cards")
	}
	ceWantCollections(t, g, me.ID, 1)

	// No targets: X = 0, nothing to exile.
	again := NewCard("Test Necropsy", me.ID)
	again.TypeLine = "Instant"
	again.OracleID = ceTargetsOracle
	me.Hand.PushTop(again)
	if err := g.CastSpell(me.ID, again.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("no targets, no evidence: %v", err)
	}
	ceWantCollections(t, g, me.ID, 2)
}

// A restore taken mid-cast keeps the pick. The test cost also returns a
// creature to hand, so naming a commander parks the cast on its owner's CR
// 903.9b question with nothing paid; the parked announcement carries the
// targets and exile_ids, and the restored answer casts the spell with the
// same evidence, checked against the same X.
func TestACollectEvidenceTargetsPickSurvivesAMidCastRestore(t *testing.T) {
	g, me, spell := ceTargetsTable(t, func(ac *AdditionalCost) {
		ac.ReturnToHand = &ReturnToHandCost{Count: 1, Filter: creatureCostSpec(), Label: "a creature you control"}
	})
	opp := g.Seats[1]
	golem := ceCard(g.Battlefield, opp.ID, "Golem", "Artifact Creature — Golem", "{3}")
	four := ceCard(me.Graveyard, me.ID, "Four", "Sorcery", "{4}")
	cmd := seedCostCard(g.Battlefield, me.ID, me.ID, true, nil)

	if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: ceRefs(golem), ExileIDs: []uuid.UUID{four}, ReturnIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	parkedCostPrompt(t, g, me.ID, cmd)
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(four) {
		t.Fatal("the cast paid before the commander's owner answered")
	}

	_, restored := roundTrip(t, g)
	rme := restored.Seats[0]
	c := parkedCostPrompt(t, restored, rme.ID, cmd)
	if err := restored.ResolveOptionalReplacement(c.ID, rme.ID, true); err != nil {
		t.Fatalf("answer on the restored table: %v", err)
	}
	if !restored.Stack.Contains(spell) || !restored.Exile.Contains(four) || rme.Graveyard.Contains(four) {
		t.Fatal("the restored answer did not cast the spell with the parked evidence")
	}
	ceWantCollections(t, restored, rme.ID, 1)
}

// The validator reads the floor from the targets and nothing else: the
// clause itself prints none, so with no targets any graveyard cards are a
// payment, and a pick that would pay X = 3 still pays it when the targets
// are worth less.
func TestCollectEvidenceTargetsAcceptsMoreThanTheFloor(t *testing.T) {
	g, me, spell := ceTargetsTable(t, nil)
	opp := g.Seats[1]
	cheap := ceCard(g.Battlefield, opp.ID, "Ornament", "Artifact", "{1}")
	four := ceCard(me.Graveyard, me.ID, "Four", "Sorcery", "{4}")
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Targets: ceRefs(cheap), ExileIDs: []uuid.UUID{four}}); err != nil {
		t.Fatalf("cards worth 4 for a target worth 1: %v", err)
	}
	if !g.Exile.Contains(four) {
		t.Error("the card was not exiled")
	}
}

// A clause read as collect evidence N is not the target-sized one, and
// the other way round, so their two readers cannot be confused.
func TestCollectEvidenceTargetsClauseIsItsOwnShape(t *testing.T) {
	sized := CollectEvidenceFromTargetsCost("x")
	if !sized.IsCollectEvidenceFromTargets() || sized.IsCollectEvidence() {
		t.Errorf("the target-sized clause reads as %v / %v", sized.IsCollectEvidenceFromTargets(), sized.IsCollectEvidence())
	}
	fixed := CollectEvidenceCost(3, "x")
	if fixed.IsCollectEvidenceFromTargets() || !fixed.IsCollectEvidence() {
		t.Error("collect evidence 3 reads as the target-sized clause")
	}
	var none *ExileCost
	if none.IsCollectEvidenceFromTargets() {
		t.Error("a nil clause reads as target-sized")
	}
}

func ceWantCollections(t *testing.T, g *Game, who uuid.UUID, n int) {
	t.Helper()
	got := 0
	for _, ev := range g.Events {
		if ev.Kind == EventCollectEvidence && ev.Actor == who {
			got++
		}
	}
	if got != n {
		t.Errorf("%d collect evidence events, want %d", got, n)
	}
}
