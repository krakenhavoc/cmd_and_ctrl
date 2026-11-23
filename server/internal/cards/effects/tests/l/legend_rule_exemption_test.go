package testsl

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

func TestMirrorBoxExemptsYourLegendsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mirror Box", "Artifact", effectstest.OracleMirrorBox, 0, 0)
	a := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	b := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	c := effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	d := effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)

	g.RunStateChecksForTest()

	if !effectstest.OnBoard(g, a) || !effectstest.OnBoard(g, b) || effectstest.LegendPrompt(g, me.ID) != nil {
		t.Error(effectstest.BothLegendsMustStay + " under your Mirror Box")
	}
	// Your static does not reach an opponent's permanents.
	ch := effectstest.LegendPrompt(g, opp.ID)
	if ch == nil {
		t.Fatal("the opponent's duplicate legends were exempted by your Mirror Box")
	}
	if len(ch.PickTargetCards) != 2 || !effectstest.OnBoard(g, c) || !effectstest.OnBoard(g, d) {
		t.Errorf("opponent prompt malformed: %v", ch.PickTargetCards)
	}
}

func TestMirrorBoxLeavingRestoresTheLegendRuleAtTheNextCheck(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	box := effectstest.B12Push(g, me.ID, "Mirror Box", "Artifact", effectstest.OracleMirrorBox, 0, 0)
	a := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	b := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) != nil {
		t.Fatal("prompted while the Box was in play")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(box); err != nil {
			t.Fatalf("destroy Box: %v", err)
		}
	})
	g.RunStateChecksForTest()

	ch := effectstest.LegendPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("no legend-rule prompt after the Box left")
	}
	if err := g.ResolveLegendRule(ch.ID, me.ID, b); err != nil {
		t.Fatalf("ResolveLegendRule: %v", err)
	}
	if effectstest.OnBoard(g, a) || !effectstest.OnBoard(g, b) {
		t.Error("the controller's choice was not honoured")
	}
}

func TestMirrorGalleryExemptsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mirror Gallery", "Artifact", effectstest.OracleMirrorGallery, 0, 0)
	ids := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2),
		effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2),
		effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2),
		effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2),
	}
	g.RunStateChecksForTest()
	for _, id := range ids {
		if !effectstest.OnBoard(g, id) {
			t.Error(effectstest.BothLegendsMustStay + " under Mirror Gallery")
		}
	}
	if effectstest.LegendPrompt(g, me.ID) != nil || effectstest.LegendPrompt(g, opp.ID) != nil {
		t.Error("a legend-rule prompt opened under Mirror Gallery")
	}
}

func TestMirrorBoxLegendsStillGetTheirAnthem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mirror Box", "Artifact", effectstest.OracleMirrorBox, 0, 0)
	a := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	// +1/+1 legendary, +1/+1 for the other same-named nontoken creature.
	if p := effectstest.EffectivePower(t, g, a); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

func TestSakashimaOfAThousandFacesExemptsItsControllerAsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sakashima of a Thousand Faces", effectstest.LegendTypeLine, effectstest.OracleSakashimaFaces, 3, 1)
	a := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	b := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	g.RunStateChecksForTest()
	if !effectstest.OnBoard(g, a) || !effectstest.OnBoard(g, b) || effectstest.LegendPrompt(g, me.ID) != nil {
		t.Error(effectstest.BothLegendsMustStay)
	}
	if effectstest.LegendPrompt(g, opp.ID) == nil {
		t.Error("opponent's legends were exempted")
	}
}

func TestSakashimaOfAThousandFacesCopyKeepsTheExemption(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.SeedCopyableCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	legend := effectstest.SeedCopyableCreature(g, active.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)

	id := effectstest.CastCatalogSpell(t, g, "Sakashima of a Thousand Faces", effectstest.LegendTypeLine, effectstest.OracleSakashimaFaces, nil)
	c := effectstest.CopyPrompt(g)
	for i := 0; c == nil && i < 8; i++ {
		_ = g.PassPriority()
		c = effectstest.CopyPrompt(g)
	}
	if c == nil {
		t.Fatal("no copy prompt")
	}
	for _, cand := range c.PickTargetCards {
		if cand == theirs {
			t.Fatal("an opponent's creature was offered as a copy source")
		}
	}
	effectstest.ResolveWithCopyChoice(t, g, legend)

	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != effectstest.LegendName {
		t.Errorf("name = %q, want the copied %q", got.Name, effectstest.LegendName)
	}
	// Two same-named legends: the original and the copy. Both stay.
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, active.ID) != nil || !effectstest.OnBoard(g, legend) || !effectstest.OnBoard(g, id) {
		t.Error("the copy lost Sakashima's exemption, so the legend rule fired")
	}
}
