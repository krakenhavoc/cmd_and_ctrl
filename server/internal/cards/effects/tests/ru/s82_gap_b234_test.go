package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b234_test.go — Liliana, Heretical Healer // Liliana, Defiant
// Necromancer (GAP batch 234, #5948), played through the real catalog entries.

const gapB234Liliana = "b96a8ad2-3d86-459b-a34f-19c9dc07c3c4"

func gapB234Row() cards.Card {
	row := effectstest.TransformRow(gapB234Liliana, "Liliana, Heretical Healer", "Legendary Creature — Human Cleric", "{1}{B}{B}",
		"Liliana, Defiant Necromancer", "Legendary Planeswalker — Liliana", "", "", []string{"B"})
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = "2", "3"
	row.CardFaces[1].Loyalty = "3"
	return row
}

func gapB234Zombies(g *game.Game, p *game.Player) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == p.ID && c.Name == "Zombie" && c.IsToken() && c.Power == 2 && c.Toughness == 2 {
			n++
		}
	}
	return n
}

func TestGapB234LilianaIsRegisteredFullOnBothFaces(t *testing.T) {
	for oracle, name := range map[string]string{gapB234Liliana: "Liliana, Heretical Healer", gapB234Liliana + "#1": "Liliana, Defiant Necromancer"} {
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

func TestGapB234LilianaFlipsAndMakesAZombieWhenAnotherNontokenCreatureOfMineDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	front := effectstest.ImportToBattlefield(t, g, gapB234Row(), me)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(front) {
		t.Fatal("Liliana flipped when a creature I do not control died")
	}
	token := effectstest.B12Creature(g, me.ID, "Token", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	// (the fixture is a nontoken card, so this death flips her)
	walker, ok := s82g58Flipped(g, me, gapB234Liliana)
	if !ok {
		t.Fatal("Liliana did not transform when another nontoken creature of mine died")
	}
	if g.Battlefield.Contains(front) {
		t.Error("the front face is still the same object")
	}
	if effectstest.E2Card(t, g, walker).Name != "Liliana, Defiant Necromancer" || effectstest.LoyaltyOf(g, walker) != 3 {
		t.Errorf("back face wrong, loyalty %d", effectstest.LoyaltyOf(g, walker))
	}
	if got := gapB234Zombies(g, me); got != 1 {
		t.Errorf("%d 2/2 Zombie tokens, want 1", got)
	}
}

func TestGapB234LilianaIgnoresMyTokensDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	front := effectstest.ImportToBattlefield(t, g, gapB234Row(), me)
	tok := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tok) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(front) {
		t.Error("Liliana flipped when a token died")
	}
}

func TestGapB234DefiantNecromancerPlusTwoMakesEachPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Liliana, Defiant Necromancer", gapB234Liliana+"#1", 3)
	for _, p := range []*game.Player{me, opp} {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		for _, p := range g.Seats {
			if c := effectstest.LatestChooseCardsFor(g, p.ID); c != nil {
				if err := g.ResolveChooseCards(c.ID, p.ID, c.ChooseCards[:1]); err != nil {
					t.Fatalf("discard: %v", err)
				}
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != mine-1 || opp.Hand.Size() != theirs-1 {
		t.Errorf("hands %d and %d, want %d and %d", me.Hand.Size(), opp.Hand.Size(), mine-1, theirs-1)
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestGapB234DefiantNecromancerMinusXReturnsANonlegendaryCreatureWithThatManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Liliana, Defiant Necromancer", gapB234Liliana+"#1", 5)
	three := effectstest.GraveyardCreature(me, "Three Drop", "Bear", 3, 3) // {2}{G}
	legend := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: legend, Name: "Legend", TypeLine: "Legendary Creature — Bear",
		ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{Strict: true,
		XValue: 2, Targets: effectstest.DgCardRef(three)}); err == nil {
		t.Fatal("a mana value 3 card was accepted for X = 2")
	}
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{Strict: true,
		XValue: 3, Targets: effectstest.DgCardRef(legend)}); err == nil {
		t.Fatal("a legendary creature card was accepted")
	}
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: effectstest.DgCardRef(three)})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) && !gapB234OnBattlefieldByName(g, me, "Three Drop") {
		t.Error("the creature card did not return")
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2 (5 − 3)", got)
	}
}

func gapB234OnBattlefieldByName(g *game.Game, p *game.Player, name string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == p.ID && c.Name == name {
			return true
		}
	}
	return false
}

func TestGapB234DefiantNecromancerEmblemReturnsEachDyingCreatureAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushWalkerForTest(g, me.ID, "Liliana, Defiant Necromancer", gapB234Liliana+"#1", 8)
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(effectstest.EmblemsOf(g, me.ID)))
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if gapB234OnBattlefieldByName(g, me, "Their Bear") {
		t.Fatal("it returned before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !gapB234OnBattlefieldByName(g, me, "Their Bear") {
		t.Error("the dead creature did not return under my control at the end step")
	}
}
