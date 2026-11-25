package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_seam_6156_cards_test.go — #6156: a search whose found cards wait,
// revealed, where they were found while an opponent (Ecological
// Appreciation) or the caster (Verdant Mastery) splits them, and a land put
// onto the battlefield under an opponent's control.

const (
	seam6156EcologicalOracle = "0635ab65-fc1a-4e43-9e11-3a1ad3093830"
	seam6156VerdantOracle    = "b392fabc-51f2-4ae4-b21c-1c662e4c6953"
)

func TestSeam6156CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		seam6156EcologicalOracle: "Ecological Appreciation",
		seam6156VerdantOracle:    "Verdant Mastery",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// seam6156Creature puts a creature card with the given mana cost into zone.
func seam6156Creature(g *game.Game, p *game.Player, zone *game.Zone, name, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		zone.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Beast", ManaCost: cost,
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	})
	return id
}

func seam6156Main(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// CR 701.23a, 701.20a: X=3 finds four creatures over the library and the
// graveyard; the opponent picks two to shuffle back, the other two enter.
func TestEcologicalAppreciationOpponentShufflesTwoBackTheRestEnter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seam6156Main(t, g)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	a := seam6156Creature(g, me, me.Library, "Alpha Beast", "{1}{G}")
	b := seam6156Creature(g, me, me.Library, "Bravo Beast", "{2}{G}")
	big := seam6156Creature(g, me, me.Library, "Huge Beast", "{4}{G}")
	twin := seam6156Creature(g, me, me.Library, "Alpha Beast", "{1}{G}")
	d := seam6156Creature(g, me, me.Graveyard, "Delta Beast", "{G}")
	e := seam6156Creature(g, me, me.Graveyard, "Echo Beast", "{2}{G}")

	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPool(me, 5)
	spell := effectstest.CastFromHandForTest(t, g, me, "Ecological Appreciation", "Sorcery", "{X}{2}{G}",
		seam6156EcologicalOracle, game.CastSpellParams{Strict: true, XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if c.SearchZonesPending() {
		t.Fatal("the search asked for zones: it names both the library and the graveyard")
	}
	for _, id := range c.SearchCards {
		if id == big {
			t.Fatal("a creature with mana value 5 was offered for X=3")
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{a, twin}); err == nil {
		t.Fatal("two creatures with the same name were found")
	}
	found := []uuid.UUID{a, b, d, e}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, found); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Graveyard.Contains(d) || !me.Library.Contains(a) {
		t.Fatal("a found card moved before the opponent chose")
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	pick := effectstest.LatestRevealPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked to choose: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(pick.ID, opp.ID, []uuid.UUID{b, d}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{b, d} {
		if !me.Library.Contains(id) {
			t.Errorf("chosen card %s was not shuffled into the library", id)
		}
	}
	for _, id := range []uuid.UUID{a, e} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("card %s did not enter the battlefield", id)
		}
	}
	if !g.Exile.Contains(spell) {
		t.Error("Ecological Appreciation did not exile itself")
	}
}

// Without the alternative cost: two lands enter tapped under the caster's
// control and the rest go to the hand; no opponent gets one.
func TestVerdantMasteryFullCostSplitsBetweenBattlefieldAndHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seam6156Main(t, g)
	lands := effectstest.SeedTypedLibrary(me, "Basic Land — Forest", "Forest", "Forest", "Forest", "Forest", "Forest")
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPool(me, 5)
	effectstest.CastFromHandForTest(t, g, me, "Verdant Mastery", "Sorcery", "{5}{G}",
		seam6156VerdantOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	found := lands[:4]
	if err := g.ResolveSearchLibrary(c.ID, me.ID, found); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	pick := effectstest.LatestRevealPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the caster was not asked which two enter: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("pick bounds %d..%d, want exactly two", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, found[:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range found[:2] {
		if !g.Battlefield.Contains(id) || !seam6156Tapped(g, id) || seam6156Controller(g, id) != me.ID {
			t.Errorf("land %s did not enter tapped under the caster's control", id)
		}
	}
	for _, id := range found[2:] {
		if !me.Hand.Contains(id) {
			t.Errorf("land %s did not go to the hand", id)
		}
	}
	if !me.Library.Contains(lands[4]) {
		t.Error("the unfound land left the library")
	}
}

// With the {3}{G} cost paid: one land enters tapped under the chosen
// opponent's control, two under the caster's, the rest to hand.
func TestVerdantMasteryAltCostGivesOneLandToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	seam6156Main(t, g)
	lands := effectstest.SeedTypedLibrary(me, "Basic Land — Forest", "Forest", "Forest", "Forest", "Forest")
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPool(me, 3)
	effectstest.CastFromHandForTest(t, g, me, "Verdant Mastery", "Sorcery", "{5}{G}",
		seam6156VerdantOracle, game.CastSpellParams{Strict: true, AlternativeCost: "verdant-mastery"})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	found := lands[:4]
	if err := g.ResolveSearchLibrary(c.ID, me.ID, found); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	pick := effectstest.LatestRevealPickFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 {
		t.Fatalf("the caster was not asked for the opponent's land: %+v", *g.PendingChoices[0])
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, found[:1]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	pick = effectstest.LatestRevealPickFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 2 {
		t.Fatalf("the caster was not asked which two enter: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, found[1:3]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(found[0]) || !seam6156Tapped(g, found[0]) || seam6156Controller(g, found[0]) != opp.ID {
		t.Error("the opponent's land did not enter tapped under their control")
	}
	for _, id := range found[1:3] {
		if !g.Battlefield.Contains(id) || !seam6156Tapped(g, id) || seam6156Controller(g, id) != me.ID {
			t.Errorf("land %s did not enter tapped under the caster's control", id)
		}
	}
	if !me.Hand.Contains(found[3]) {
		t.Error("the last land did not go to the hand")
	}
}

func seam6156Tapped(g *game.Game, id uuid.UUID) bool {
	tapped := false
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			tapped = c.Tapped
		}
	})
	return tapped
}

func seam6156Controller(g *game.Game, id uuid.UUID) uuid.UUID {
	var who uuid.UUID
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			who = c.Controller
		}
	})
	return who
}
