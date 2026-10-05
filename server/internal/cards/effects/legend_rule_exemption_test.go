package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// legend_rule_exemption_test.go — "the legend rule doesn't apply"
// (CR 704.5j, #2177): Mirror Box, Mirror Gallery, Sakashima of a
// Thousand Faces.

const (
	oracleMirrorBox      = "3bed1944-58dc-4679-9aee-7be4d94fb55c"
	oracleMirrorGallery  = "7c9c3060-ca9d-4868-84c6-2a95b0fa8885"
	oracleSakashimaFaces = "8ecdaf4b-4442-42da-9714-4257a83faf50"
	legendTypeLine       = "Legendary Creature — Human"
	legendName           = "Test Legend"
	bothLegendsMustStay  = "both legends must stay"
)

func legendPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceLegendRule && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func onBoard(g *game.Game, id uuid.UUID) bool {
	_, ok := findBattlefieldByID(g, id)
	return ok
}

func TestMirrorBoxExemptsYourLegendsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Mirror Box", "Artifact", oracleMirrorBox, 0, 0)
	a := b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	b := b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	c := b12Creature(g, opp.ID, legendName, legendTypeLine, 2, 2)
	d := b12Creature(g, opp.ID, legendName, legendTypeLine, 2, 2)

	g.RunStateChecksForTest()

	if !onBoard(g, a) || !onBoard(g, b) || legendPrompt(g, me.ID) != nil {
		t.Error(bothLegendsMustStay + " under your Mirror Box")
	}
	// Your static does not reach an opponent's permanents.
	ch := legendPrompt(g, opp.ID)
	if ch == nil {
		t.Fatal("the opponent's duplicate legends were exempted by your Mirror Box")
	}
	if len(ch.PickTargetCards) != 2 || !onBoard(g, c) || !onBoard(g, d) {
		t.Errorf("opponent prompt malformed: %v", ch.PickTargetCards)
	}
}

func TestMirrorBoxLeavingRestoresTheLegendRuleAtTheNextCheck(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	box := b12Push(g, me.ID, "Mirror Box", "Artifact", oracleMirrorBox, 0, 0)
	a := b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	b := b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	g.RunStateChecksForTest()
	if legendPrompt(g, me.ID) != nil {
		t.Fatal("prompted while the Box was in play")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(box); err != nil {
			t.Fatalf("destroy Box: %v", err)
		}
	})
	g.RunStateChecksForTest()

	ch := legendPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("no legend-rule prompt after the Box left")
	}
	if err := g.ResolveLegendRule(ch.ID, me.ID, b); err != nil {
		t.Fatalf("ResolveLegendRule: %v", err)
	}
	if onBoard(g, a) || !onBoard(g, b) {
		t.Error("the controller's choice was not honoured")
	}
}

func TestMirrorGalleryExemptsEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Mirror Gallery", "Artifact", oracleMirrorGallery, 0, 0)
	ids := []uuid.UUID{
		b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2),
		b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2),
		b12Creature(g, opp.ID, legendName, legendTypeLine, 2, 2),
		b12Creature(g, opp.ID, legendName, legendTypeLine, 2, 2),
	}
	g.RunStateChecksForTest()
	for _, id := range ids {
		if !onBoard(g, id) {
			t.Error(bothLegendsMustStay + " under Mirror Gallery")
		}
	}
	if legendPrompt(g, me.ID) != nil || legendPrompt(g, opp.ID) != nil {
		t.Error("a legend-rule prompt opened under Mirror Gallery")
	}
}

func TestMirrorBoxLegendsStillGetTheirAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Mirror Box", "Artifact", oracleMirrorBox, 0, 0)
	a := b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	// +1/+1 legendary, +1/+1 for the other same-named nontoken creature.
	if p := effectivePower(t, g, a); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

func TestSakashimaOfAThousandFacesExemptsItsControllerAsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sakashima of a Thousand Faces", legendTypeLine, oracleSakashimaFaces, 3, 1)
	a := b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	b := b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	b12Creature(g, opp.ID, legendName, legendTypeLine, 2, 2)
	b12Creature(g, opp.ID, legendName, legendTypeLine, 2, 2)
	g.RunStateChecksForTest()
	if !onBoard(g, a) || !onBoard(g, b) || legendPrompt(g, me.ID) != nil {
		t.Error(bothLegendsMustStay)
	}
	if legendPrompt(g, opp.ID) == nil {
		t.Error("opponent's legends were exempted")
	}
}

func TestSakashimaOfAThousandFacesCopyKeepsTheExemption(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := seedCopyableCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	legend := seedCopyableCreature(g, active.ID, legendName, legendTypeLine, 2, 2)

	id := castCatalogSpell(t, g, "Sakashima of a Thousand Faces", legendTypeLine, oracleSakashimaFaces, nil)
	c := copyPrompt(g)
	for i := 0; c == nil && i < 8; i++ {
		_ = g.PassPriority()
		c = copyPrompt(g)
	}
	if c == nil {
		t.Fatal("no copy prompt")
	}
	for _, cand := range c.PickTargetCards {
		if cand == theirs {
			t.Fatal("an opponent's creature was offered as a copy source")
		}
	}
	resolveWithCopyChoice(t, g, legend)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != legendName {
		t.Errorf("name = %q, want the copied %q", got.Name, legendName)
	}
	// Two same-named legends: the original and the copy. Both stay.
	g.RunStateChecksForTest()
	if legendPrompt(g, active.ID) != nil || !onBoard(g, legend) || !onBoard(g, id) {
		t.Error("the copy lost Sakashima's exemption, so the legend rule fired")
	}
}
