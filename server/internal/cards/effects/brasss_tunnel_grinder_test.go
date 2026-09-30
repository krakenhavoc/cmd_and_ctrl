package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const brassTunnelGrinderOracle = "af1553eb-4f9f-4335-9078-56649bd8d8fc"

// btgCard is the double-faced card as the deck importer builds it,
// front face up.
func btgCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   brassTunnelGrinderOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Brass's Tunnel-Grinder", TypeLine: "Legendary Artifact", ManaCost: "{2}{R}", Colors: []string{"R"}},
			{Name: "Tecutlan, the Searing Rift", TypeLine: "Legendary Land — Cave"},
		},
	}
	c.SetFace(0)
	return c
}

// The ETB: discard any number, then draw that many plus one.
func TestBrassTunnelGrinderDiscardsAnyNumberThenDrawsThatManyPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	a := ssCard(me.Hand, me.ID, "Card A", "Sorcery", 0, 0)
	b := ssCard(me.Hand, me.ID, "Card B", "Instant", 0, 0)
	ssCard(me.Hand, me.ID, "Card C", "Instant", 0, 0)

	castCatalogSpell(t, g, "Brass's Tunnel-Grinder", "Legendary Artifact", brassTunnelGrinderOracle, nil)
	passPriorityAroundTable(t, g)

	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the ETB asks which cards to discard")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 3 {
		t.Errorf("any number: bounds [%d,%d], want [0,3]", c.ChooseMin, c.ChooseMax)
	}
	answerDiscard(t, g, me.ID, a, b)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Fatal("the two chosen cards are discarded")
	}
	if got := me.Hand.Size(); got != 1+3 {
		t.Errorf("hand %d, want 4 (one kept, three drawn for two discarded)", got)
	}
}

// Discarding nothing still draws one.
func TestBrassTunnelGrinderDiscardingNothingDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	ssCard(me.Hand, me.ID, "Card A", "Sorcery", 0, 0)

	castCatalogSpell(t, g, "Brass's Tunnel-Grinder", "Legendary Artifact", brassTunnelGrinderOracle, nil)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2 (kept one, drew one)", got)
	}
}

// With an empty hand there is nothing to ask, and it still draws one.
func TestBrassTunnelGrinderWithAnEmptyHandDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil

	castCatalogSpell(t, g, "Brass's Tunnel-Grinder", "Legendary Artifact", brassTunnelGrinderOracle, nil)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("an empty hand is asked nothing")
	}
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1", got)
	}
}

// Descending gives a bore counter at the end step; the third transforms
// it, in place, into Tecutlan, and takes the counters off.
func TestBrassTunnelGrinderTransformsOnTheThirdBoreCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := btgCard(me.ID)
	card.Counters = map[string]int{"bore": 2}
	grinder := pushBattlefieldCardWithTimestamp(g, card)

	// Descend: a permanent card milled into your graveyard.
	ssCard(me.Library, me.ID, "Milled Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	got, ok := battlefieldCard(g, grinder)
	if !ok {
		t.Fatal("the permanent is still on the battlefield (transformed in place)")
	}
	if got.Name != "Tecutlan, the Searing Rift" || !got.IsLand() {
		t.Errorf("it transformed into Tecutlan: %q %q", got.Name, got.TypeLine)
	}
	if n := got.Counters["bore"]; n != 0 {
		t.Errorf("the bore counters are removed: %d", n)
	}
}

// Without descending there is no trigger at all (intervening if).
func TestBrassTunnelGrinderNeedsYouToHaveDescended(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grinder := pushBattlefieldCardWithTimestamp(g, btgCard(me.ID))

	// An instant card is not a permanent card: no descent.
	ssCard(me.Library, me.ID, "Milled Bolt", "Instant", 0, 0)
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	advanceToEndStepOf(t, g, 0)
	if triggersOnStackFrom(g, grinder) != 0 {
		t.Fatal("no descent, no trigger")
	}
	passPriorityAroundTable(t, g)
	got, _ := battlefieldCard(g, grinder)
	if n := got.Counters["bore"]; n != 0 {
		t.Errorf("no bore counter without descending: %d", n)
	}
}

// One bore counter below three leaves the front face up.
func TestBrassTunnelGrinderFirstBoreCounterDoesNotTransform(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grinder := pushBattlefieldCardWithTimestamp(g, btgCard(me.ID))
	ssCard(me.Library, me.ID, "Milled Land", "Basic Land — Mountain", 0, 0)
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	got, _ := battlefieldCard(g, grinder)
	if n := got.Counters["bore"]; n != 1 {
		t.Errorf("one bore counter: %d", n)
	}
	if got.Name != "Brass's Tunnel-Grinder" {
		t.Errorf("still the front face: %q", got.Name)
	}
}

// Tecutlan's mana spent on a permanent spell discovers X, X being that
// spell's mana value; spent on an instant it does nothing (ADR 0099).
func TestTecutlanDiscoversOffAPermanentSpellItsManaPaysFor(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		discovers      bool
	}{
		{"a creature spell", "Creature — Goblin", true},
		{"an instant", "Instant", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			hit := discoverLibraryCard(me.ID, "Shock", "Instant", "{R}")
			me.Library.Cards = []game.Card{hit}
			card := btgCard(me.ID)
			card.SetFace(1)
			land := pushBattlefieldCardWithTimestamp(g, card)
			activateManaFor(t, g, me.ID, land, 0, game.ManaAbilityParams{})

			castFromHandForTest(t, g, me, "Spell", tc.typeLine, "{R}", "", strict)
			passPriorityAroundTable(t, g)
			var prompt *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceMayCast {
					prompt = c
				}
			}
			if !tc.discovers {
				if prompt != nil || !me.Library.Contains(hit.InstanceID) {
					t.Fatalf("an instant paid with Tecutlan's mana discovered")
				}
				return
			}
			if prompt == nil || prompt.MayCastCard != hit.InstanceID {
				t.Fatalf("discover 1 did not offer the one-drop: %+v", prompt)
			}
		})
	}
}

// Tecutlan taps for {R}.
func TestTecutlanTapsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := btgCard(me.ID)
	card.SetFace(1)
	land := pushBattlefieldCardWithTimestamp(g, card)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	red := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "R" {
			red++
		}
	}
	if red != 1 || len(me.ManaPool) != 1 {
		t.Errorf("pool %+v, want one red mana", me.ManaPool)
	}
}
