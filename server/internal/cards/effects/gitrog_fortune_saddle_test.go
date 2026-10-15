package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gitrog_fortune_saddle_test.go — #3547, ADR 0186 slice 5: the two cards
// that read "a creature that saddled it this turn" after the saddle
// record landed (slice 1).

const (
	gitrogRavenousRideOracle = "7739311a-682b-48a6-8db2-c07f44141751"
	fortuneLoyalSteedOracle  = "dc440b62-0bf0-4f23-9210-3550ff0a8d4d"
)

// pushHandLands puts n basic lands into the player's hand and returns them.
func pushHandLands(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Test Forest", TypeLine: "Basic Land — Forest",
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	return ids
}

// gitrogConnects saddles The Gitrog with `saddler`, attacks the second
// seat and runs combat damage until the sacrifice prompt is up.
func gitrogConnects(t *testing.T, g *game.Game, me, opp *game.Player, gitrog, saddler uuid.UUID) *game.PendingChoice {
	t.Helper()
	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, gitrog, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, gitrog)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	choice := latestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt for the creature that saddled The Gitrog")
	}
	return choice
}

func TestGitrogRavenousRideSacrificesTheSaddlerDrawsXAndPutsLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := pushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", gitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 3)
	lands := pushHandLands(me, 4)
	handBefore := me.Hand.Size()
	oppLife := opp.Life

	choice := gitrogConnects(t, g, me, opp, gitrog, saddler)
	if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != saddler {
		t.Fatalf("offered %v, want exactly the saddler", choice.ChooseCards)
	}
	if opp.Life != oppLife-6 {
		t.Errorf("defender at %d, want %d", opp.Life, oppLife-6)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	if !me.Graveyard.Contains(saddler) {
		t.Fatal("the saddler was not sacrificed")
	}
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand %d, want %d (drew X = 3)", got, handBefore+3)
	}
	put := latestChooseCards(g, me.ID)
	if put == nil {
		t.Fatal("no prompt for the lands")
	}
	if put.ChooseMax != 3 {
		t.Errorf("the put offers up to %d, want 3", put.ChooseMax)
	}
	if err := g.ResolveChooseCards(put.ID, me.ID, lands[:3]); err != nil {
		t.Fatalf("put lands: %v", err)
	}
	for _, id := range lands[:3] {
		c, ok := battlefieldCardByID(g, id)
		if !ok {
			t.Errorf("land %s did not enter", id)
		} else if !c.Tapped {
			t.Errorf("land %s entered untapped", id)
		}
	}
	if !me.Hand.Contains(lands[3]) {
		t.Error("the fourth land left the hand")
	}
}

func TestGitrogRavenousRideDecliningSacrificesAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := pushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", gitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 3)
	handBefore := me.Hand.Size()

	choice := gitrogConnects(t, g, me, opp, gitrog, saddler)
	if err := g.ResolveChooseCards(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if _, ok := battlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler was sacrificed though the player declined")
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d, want %d (nothing drawn)", got, handBefore)
	}
	if latestChooseCards(g, me.ID) != nil {
		t.Error("declining still queued a land prompt")
	}
}

// X is the power the creature had as it was sacrificed: an until-end-of-turn
// pump counts.
func TestGitrogRavenousRideCountsAPumpInX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := pushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", gitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 3)
	handBefore := me.Hand.Size()

	choice := gitrogConnects(t, g, me, opp, gitrog, saddler)
	g.WithWriteLock(func() {
		if err := (BoostUntilEOT{Target: saddler, Power: 2, Toughness: 2}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("boost: %v", err)
		}
	})
	// Reading the effective power settles the layer cache, as the state
	// check after a real pump would.
	if got := effectivePower(t, g, saddler); got != 5 {
		t.Fatalf("pumped saddler has power %d, want 5", got)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	if got := me.Hand.Size(); got != handBefore+5 {
		t.Errorf("hand %d, want %d (3 printed + 2 pump = 5 cards)", got, handBefore+5)
	}
}

// A creature that did not saddle The Gitrog is not offered.
func TestGitrogRavenousRideOffersOnlyTheSaddlers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := pushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", gitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 3)
	bystander := pushCrewerForTest(g, me.ID, "Bystander", 4)

	choice := gitrogConnects(t, g, me, opp, gitrog, saddler)
	for _, id := range choice.ChooseCards {
		if id == bystander {
			t.Errorf("the bystander %s was offered", bystander)
		}
	}
}

// fortuneAttacks saddles Fortune with the given creatures and runs the
// attack, so the delayed trigger is scheduled and combat is about to
// start dealing damage.
func fortuneAttacks(t *testing.T, g *game.Game, me, opp *game.Player, fortune uuid.UUID, saddlers ...uuid.UUID) {
	t.Helper()
	toMain(t, g)
	if err := saddleAbility(t, g, me.ID, fortune, saddlers...); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, fortune)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
}

// settleFortuneReturn runs the end of combat to the point the choice is up.
func fortuneEndOfCombatPrompt(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	choice := latestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt for the creature that saddled Fortune")
	}
	return choice
}

// Fortune's returned ETB scries; answer it if it is up.
func answerScryIfAsked(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		answerScryKeepAll(t, g, me.ID)
	}
}

func TestFortuneFlickersItselfAndTheChosenSaddlerAtEndOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := pushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", fortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)

	fortuneAttacks(t, g, me, opp, fortune, saddler)
	choice := fortuneEndOfCombatPrompt(t, g, me)
	if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != saddler {
		t.Fatalf("offered %v, want exactly the saddler", choice.ChooseCards)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	answerScryIfAsked(t, g, me)

	for _, old := range []uuid.UUID{fortune, saddler} {
		if _, ok := battlefieldCardByID(g, old); ok {
			t.Errorf("%s is still the same object on the battlefield", old)
		}
	}
	if ids := battlefieldIDsNamed(g, "Fortune, Loyal Steed"); len(ids) != 1 {
		t.Errorf("%d Fortunes on the battlefield, want 1 (returned)", len(ids))
	}
	if ids := battlefieldIDsNamed(g, "Saddler"); len(ids) != 1 {
		t.Errorf("%d Saddlers on the battlefield, want 1 (returned)", len(ids))
	}
}

// "Up to one": choosing none flickers only Fortune.
func TestFortuneChoosingNoneFlickersOnlyFortune(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := pushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", fortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)

	fortuneAttacks(t, g, me, opp, fortune, saddler)
	choice := fortuneEndOfCombatPrompt(t, g, me)
	if err := g.ResolveChooseCards(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("choose none: %v", err)
	}
	answerScryIfAsked(t, g, me)

	if _, ok := battlefieldCardByID(g, fortune); ok {
		t.Error("Fortune was not flickered")
	}
	if ids := battlefieldIDsNamed(g, "Fortune, Loyal Steed"); len(ids) != 1 {
		t.Errorf("%d Fortunes on the battlefield, want 1", len(ids))
	}
	if _, ok := battlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler was flickered though none was chosen")
	}
}

// The ruling: the delayed trigger works whether or not Fortune is on the
// battlefield, so its controller can still exile a creature that saddled it.
func TestFortuneKilledBeforeEndOfCombatStillFlickersASaddler(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := pushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", fortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)

	fortuneAttacks(t, g, me, opp, fortune, saddler)
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(fortune); err != nil {
			t.Fatalf("killing Fortune: %v", err)
		}
	})
	if !me.Graveyard.Contains(fortune) {
		t.Fatal("Fortune did not die")
	}
	choice := fortuneEndOfCombatPrompt(t, g, me)
	if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != saddler {
		t.Fatalf("offered %v, want the saddler of the dead Fortune", choice.ChooseCards)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if _, ok := battlefieldCardByID(g, saddler); ok {
		t.Error("the saddler was not flickered")
	}
	if ids := battlefieldIDsNamed(g, "Saddler"); len(ids) != 1 {
		t.Errorf("%d Saddlers on the battlefield, want 1 (returned)", len(ids))
	}
	if !me.Graveyard.Contains(fortune) {
		t.Error("the dead Fortune left the graveyard")
	}
}

// A token that saddled it is exiled and does not come back (CR 111.8).
func TestFortuneTokenSaddlerDoesNotReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := pushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", fortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	token := pushDiesCreatureForTest(g, me.ID, "Test Token", "", "Token Creature — Test", 1, 1)

	fortuneAttacks(t, g, me, opp, fortune, token)
	choice := fortuneEndOfCombatPrompt(t, g, me)
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{token}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	answerScryIfAsked(t, g, me)

	if ids := battlefieldIDsNamed(g, "Test Token"); len(ids) != 0 {
		t.Errorf("the token came back (%d on the battlefield)", len(ids))
	}
	if ids := battlefieldIDsNamed(g, "Fortune, Loyal Steed"); len(ids) != 1 {
		t.Errorf("%d Fortunes on the battlefield, want 1", len(ids))
	}
}
