package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// With nothing in exile, Coax is a wish for an Eldrazi card.
func TestCoaxWithNothingInExileIsAWishForAnEldrazi(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Ulamog's Crusher", "Creature — Eldrazi"), effectstest.OutsideCard("Grizzly Bears", "Creature — Bear"))
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", effectstest.CoaxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || c.ChooseDestination != game.ChooseFromOutside {
		t.Fatalf("no outside pick: %+v", g.PendingChoices)
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Ulamog's Crusher"}) {
		t.Errorf("candidates %v, want the Eldrazi alone", got)
	}
	crusher := effectstest.SideCardID(me, "Ulamog's Crusher")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{crusher}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, crusher) {
		t.Fatal("the Eldrazi did not come into the hand")
	}
	for _, card := range me.Hand.Cards {
		if card.InstanceID == crusher && !card.IsKnownTo(opp.ID) {
			t.Error("the card from outside the game was not revealed")
		}
	}
}

// Coax offers only the caster's own face-up Eldrazi cards in exile, and
// taking one ends the spell: nothing is asked about outside the game.
func TestCoaxTakesAFaceUpEldraziYouOwnFromExile(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Ulamog's Crusher", "Creature — Eldrazi"))
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.CoaxExile(g, me.ID, "Eldrazi Devastator", "Creature — Eldrazi", false)
	effectstest.CoaxExile(g, me.ID, "Hidden Eldrazi", "Creature — Eldrazi", true)
	effectstest.CoaxExile(g, opp.ID, "Their Eldrazi", "Creature — Eldrazi", false)
	effectstest.CoaxExile(g, me.ID, "Grizzly Bears", "Creature — Bear", false)

	effectstest.CastCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", effectstest.CoaxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || c.ChooseDestination == game.ChooseFromOutside {
		t.Fatalf("no exile pick first: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != mine || c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Fatalf("exile candidates %v (%d..%d), want my face-up Eldrazi alone, optional", c.ChooseCards, c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, mine) {
		t.Error("the exiled Eldrazi did not come into the hand")
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("taking a card from exile went on to ask about outside the game")
	}
	if len(me.SideboardCards()) != 1 {
		t.Error("the sideboard was touched")
	}
}

// Declining the exile pick is choosing to look outside the game instead,
// here in a deck with no sideboard: the card database.
func TestCoaxDecliningExileLooksOutsideTheGame(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	effectstest.CoaxExile(g, me.ID, "Eldrazi Devastator", "Creature — Eldrazi", false)
	effectstest.CastCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", effectstest.CoaxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no exile pick")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.LatestCollectionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("declining exile did not look outside the game: %+v", g.PendingChoices)
	}
	if !effectstest.EqualNames(pick.CollectionSuggestions, []string{"Ulamog's Crusher"}) {
		t.Errorf("suggestions %v, want the database's Eldrazi", pick.CollectionSuggestions)
	}
	if err := g.ResolveOutsideCollection(pick.ID, me.ID, []string{"Ulamog's Crusher"}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.HandCardNamed(me, "Ulamog's Crusher"); !ok {
		t.Error("the Eldrazi from the database did not come into the hand")
	}
}

// The static: an opponent's artifact cannot activate even a mana
// ability; Karn's controller's own can.
func TestKarnStopsYourOpponentsArtifactsOnly(t *testing.T) {
	g, _ := effectstest.KarnGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}", OracleID: effectstest.SolRingOracle,
		Owner: them.ID, Controller: them.ID,
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}", OracleID: effectstest.SolRingOracle,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(them.ID, theirs, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the opponent's Sol Ring: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("Karn's controller's own Sol Ring: err = %v", err)
	}
}

// The +1: the artifact is an artifact creature with base power and
// toughness equal to its mana value until Karn's controller's next turn.
func TestKarnsPlusOneAnimatesAnArtifactUntilYourNextTurn(t *testing.T) {
	g, karn := effectstest.KarnGame(t)
	them := g.Seats[1]
	relic := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Relic", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: them.ID, Controller: them.ID,
	})
	effectstest.Pr7Activate(t, g, g.Seats[0].ID, karn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}}})
	check := func(when string, want bool) {
		t.Helper()
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := effectstest.FindBattlefieldCardForTest(g, relic)
		if c == nil {
			t.Fatalf("%s: the relic is gone", when)
		}
		eff := c.Effective()
		if c.IsCreature() != want || !c.IsArtifact() {
			t.Fatalf("%s: creature %v artifact %v, want creature %v", when, c.IsCreature(), c.IsArtifact(), want)
		}
		if want && (eff.Power != 3 || eff.Toughness != 3) {
			t.Errorf("%s: %d/%d, want 3/3 (its mana value)", when, eff.Power, eff.Toughness)
		}
	}
	check("as it resolves", true)
	if got := effectstest.FindBattlefieldCardForTest(g, karn).Counters[game.CounterLoyalty]; got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	check("on the opponent's turn", true)
	effectstest.AdvanceToMainOf(t, g, 0)
	check("on Karn's controller's next turn", false)
}

// A creature artifact is not a legal target for the +1.
func TestKarnsPlusOneRefusesAnArtifactCreature(t *testing.T) {
	g, karn := effectstest.KarnGame(t)
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{4}",
		Power: 3, Toughness: 3, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	err := g.ActivateCatalogAbility(g.Seats[0].ID, karn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}}})
	if err == nil {
		t.Fatal("the +1 targeted an artifact creature")
	}
}

// The −2 takes a face-up artifact card its controller owns from exile.
func TestKarnsMinusTwoTakesAnArtifactFromExile(t *testing.T) {
	g, karn := effectstest.KarnGame(t)
	me := g.Seats[0]
	relic := effectstest.CoaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	effectstest.CoaxExile(g, me.ID, "Exiled Bear", "Creature — Bear", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != relic {
		t.Fatalf("exile pick %+v, want the artifact alone", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, relic) {
		t.Error("the artifact did not come into the hand")
	}
	if got := effectstest.FindBattlefieldCardForTest(g, karn).Counters[game.CounterLoyalty]; got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

// …and from outside the game when exile has none: here the database.
func TestKarnsMinusTwoTakesAnArtifactFromOutsideTheGame(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g, karn := effectstest.KarnGame(t)
	me := g.Seats[0]
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestCollectionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick from outside the game: %+v", g.PendingChoices)
	}
	if !effectstest.EqualNames(pick.CollectionSuggestions, []string{"Sol Ring"}) {
		t.Errorf("suggestions %v, want the database's artifact", pick.CollectionSuggestions)
	}
	if err := g.ResolveOutsideCollection(pick.ID, me.ID, []string{"Sol Ring"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := effectstest.HandCardNamed(me, "Sol Ring"); !ok {
		t.Error("Sol Ring did not come into the hand")
	}
}

func TestCoaxAndKarnAreFull(t *testing.T) {
	for _, id := range []string{effectstest.CoaxOracle, effectstest.KarnOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is %v with caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
