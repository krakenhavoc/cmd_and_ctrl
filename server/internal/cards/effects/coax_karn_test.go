package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// coax_karn_test.go — ADR 0198: the two cards that look in exile as well
// as outside the game, Coax from the Blind Eternities and Karn, the Great
// Creator.

const (
	coaxOracle = "63488c95-c249-40d8-a579-3ca311ed57ca"
	karnOracle = "a20dd48d-d344-4db1-b0e9-a2b71c3cc9d1"
)

// coaxExile puts a card owned by `owner` into exile.
func coaxExile(g *game.Game, owner uuid.UUID, name, typeLine string, faceDown bool) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner, FaceDown: faceDown,
	}
	g.Exile.PushTop(c)
	return c.InstanceID
}

func handHasID(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// With nothing in exile, Coax is a wish for an Eldrazi card.
func TestCoaxWithNothingInExileIsAWishForAnEldrazi(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Ulamog's Crusher", "Creature — Eldrazi"), outsideCard("Grizzly Bears", "Creature — Bear"))
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", coaxOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || c.ChooseDestination != game.ChooseFromOutside {
		t.Fatalf("no outside pick: %+v", g.PendingChoices)
	}
	if got := candidateNames(g, c); !equalNames(got, []string{"Ulamog's Crusher"}) {
		t.Errorf("candidates %v, want the Eldrazi alone", got)
	}
	crusher := sideCardID(me, "Ulamog's Crusher")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{crusher}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !handHasID(me, crusher) {
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
	g := newSideboardCatalogGame(t, 2, outsideCard("Ulamog's Crusher", "Creature — Eldrazi"))
	me, opp := g.Seats[0], g.Seats[1]
	mine := coaxExile(g, me.ID, "Eldrazi Devastator", "Creature — Eldrazi", false)
	coaxExile(g, me.ID, "Hidden Eldrazi", "Creature — Eldrazi", true)
	coaxExile(g, opp.ID, "Their Eldrazi", "Creature — Eldrazi", false)
	coaxExile(g, me.ID, "Grizzly Bears", "Creature — Bear", false)

	castCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", coaxOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || c.ChooseDestination == game.ChooseFromOutside {
		t.Fatalf("no exile pick first: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != mine || c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Fatalf("exile candidates %v (%d..%d), want my face-up Eldrazi alone, optional", c.ChooseCards, c.ChooseMin, c.ChooseMax)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !handHasID(me, mine) {
		t.Error("the exiled Eldrazi did not come into the hand")
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("taking a card from exile went on to ask about outside the game")
	}
	if len(me.SideboardCards()) != 1 {
		t.Error("the sideboard was touched")
	}
}

// Declining the exile pick is choosing to look outside the game instead,
// here in a deck with no sideboard: the card database.
func TestCoaxDecliningExileLooksOutsideTheGame(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	coaxExile(g, me.ID, "Eldrazi Devastator", "Creature — Eldrazi", false)
	castCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", coaxOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no exile pick")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	pick := latestCollectionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("declining exile did not look outside the game: %+v", g.PendingChoices)
	}
	if !equalNames(pick.CollectionSuggestions, []string{"Ulamog's Crusher"}) {
		t.Errorf("suggestions %v, want the database's Eldrazi", pick.CollectionSuggestions)
	}
	if err := g.ResolveOutsideCollection(pick.ID, me.ID, []string{"Ulamog's Crusher"}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := handCardNamed(me, "Ulamog's Crusher"); !ok {
		t.Error("the Eldrazi from the database did not come into the hand")
	}
}

// karnGame seats Karn for seat 0 in its main phase, with five loyalty.
func karnGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMainOf(t, g, 0)
	karn := pushCatalogPermanent(g, g.Seats[0].ID, "Karn, the Great Creator", "Legendary Planeswalker — Karn", karnOracle, false)
	findBattlefieldCardForTest(g, karn).Counters = map[string]int{game.CounterLoyalty: 5}
	return g, karn
}

// The static: an opponent's artifact cannot activate even a mana
// ability; Karn's controller's own can.
func TestKarnStopsYourOpponentsArtifactsOnly(t *testing.T) {
	g, _ := karnGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}", OracleID: solRingOracle,
		Owner: them.ID, Controller: them.ID,
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}", OracleID: solRingOracle,
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
	g, karn := karnGame(t)
	them := g.Seats[1]
	relic := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Relic", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: them.ID, Controller: them.ID,
	})
	pr7Activate(t, g, g.Seats[0].ID, karn, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}}})
	check := func(when string, want bool) {
		t.Helper()
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := findBattlefieldCardForTest(g, relic)
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
	if got := findBattlefieldCardForTest(g, karn).Counters[game.CounterLoyalty]; got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
	advanceToMainOf(t, g, 1)
	check("on the opponent's turn", true)
	advanceToMainOf(t, g, 0)
	check("on Karn's controller's next turn", false)
}

// A creature artifact is not a legal target for the +1.
func TestKarnsPlusOneRefusesAnArtifactCreature(t *testing.T) {
	g, karn := karnGame(t)
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
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
	g, karn := karnGame(t)
	me := g.Seats[0]
	relic := coaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	coaxExile(g, me.ID, "Exiled Bear", "Creature — Bear", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != relic {
		t.Fatalf("exile pick %+v, want the artifact alone", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !handHasID(me, relic) {
		t.Error("the artifact did not come into the hand")
	}
	if got := findBattlefieldCardForTest(g, karn).Counters[game.CounterLoyalty]; got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

// …and from outside the game when exile has none: here the database.
func TestKarnsMinusTwoTakesAnArtifactFromOutsideTheGame(t *testing.T) {
	installTestCollection(t)
	g, karn := karnGame(t)
	me := g.Seats[0]
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	pick := latestCollectionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick from outside the game: %+v", g.PendingChoices)
	}
	if !equalNames(pick.CollectionSuggestions, []string{"Sol Ring"}) {
		t.Errorf("suggestions %v, want the database's artifact", pick.CollectionSuggestions)
	}
	if err := g.ResolveOutsideCollection(pick.ID, me.ID, []string{"Sol Ring"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := handCardNamed(me, "Sol Ring"); !ok {
		t.Error("Sol Ring did not come into the hand")
	}
}

func TestCoaxAndKarnAreFull(t *testing.T) {
	for _, id := range []string{coaxOracle, karnOracle} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not in the catalog", id)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is %v with caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
