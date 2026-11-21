package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestGitrogRavenousRideSacrificesTheSaddlerDrawsXAndPutsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := effectstest.PushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", effectstest.GitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 3)
	lands := effectstest.PushHandLands(me, 4)
	handBefore := me.Hand.Size()
	oppLife := opp.Life

	choice := effectstest.GitrogConnects(t, g, me, opp, gitrog, saddler)
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
	put := effectstest.LatestChooseCards(g, me.ID)
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
		c, ok := effectstest.BattlefieldCardByID(g, id)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := effectstest.PushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", effectstest.GitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 3)
	handBefore := me.Hand.Size()

	choice := effectstest.GitrogConnects(t, g, me, opp, gitrog, saddler)
	if err := g.ResolveChooseCards(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if _, ok := effectstest.BattlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler was sacrificed though the player declined")
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d, want %d (nothing drawn)", got, handBefore)
	}
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("declining still queued a land prompt")
	}
}

// X is the power the creature had as it was sacrificed: an until-end-of-turn
// pump counts.
func TestGitrogRavenousRideCountsAPumpInX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := effectstest.PushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", effectstest.GitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 3)
	handBefore := me.Hand.Size()

	choice := effectstest.GitrogConnects(t, g, me, opp, gitrog, saddler)
	g.WithWriteLock(func() {
		if err := (effects.BoostUntilEOT{Target: saddler, Power: 2, Toughness: 2}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("boost: %v", err)
		}
	})
	// Reading the effective power settles the layer cache, as the state
	// check after a real pump would.
	if got := effectstest.EffectivePower(t, g, saddler); got != 5 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gitrog := effectstest.PushDiesCreatureForTest(g, me.ID, "The Gitrog, Ravenous Ride", effectstest.GitrogRavenousRideOracle, "Legendary Creature — Frog Horror Mount", 6, 5)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 3)
	bystander := effectstest.PushCrewerForTest(g, me.ID, "Bystander", 4)

	choice := effectstest.GitrogConnects(t, g, me, opp, gitrog, saddler)
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
	effectstest.ToMain(t, g)
	if err := effectstest.SaddleAbility(t, g, me.ID, fortune, saddlers...); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, fortune)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
}

// settleFortuneReturn runs the end of combat to the point the choice is up.
func fortuneEndOfCombatPrompt(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt for the creature that saddled Fortune")
	}
	return choice
}

func TestFortuneFlickersItselfAndTheChosenSaddlerAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := effectstest.PushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", effectstest.FortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)

	fortuneAttacks(t, g, me, opp, fortune, saddler)
	choice := fortuneEndOfCombatPrompt(t, g, me)
	if len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != saddler {
		t.Fatalf("offered %v, want exactly the saddler", choice.ChooseCards)
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{saddler}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.AnswerScryIfAsked(t, g, me)

	for _, old := range []uuid.UUID{fortune, saddler} {
		if _, ok := effectstest.BattlefieldCardByID(g, old); ok {
			t.Errorf("%s is still the same object on the battlefield", old)
		}
	}
	if ids := effectstest.BattlefieldIDsNamed(g, "Fortune, Loyal Steed"); len(ids) != 1 {
		t.Errorf("%d Fortunes on the battlefield, want 1 (returned)", len(ids))
	}
	if ids := effectstest.BattlefieldIDsNamed(g, "Saddler"); len(ids) != 1 {
		t.Errorf("%d Saddlers on the battlefield, want 1 (returned)", len(ids))
	}
}

// "Up to one": choosing none flickers only Fortune.
func TestFortuneChoosingNoneFlickersOnlyFortune(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := effectstest.PushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", effectstest.FortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)

	fortuneAttacks(t, g, me, opp, fortune, saddler)
	choice := fortuneEndOfCombatPrompt(t, g, me)
	if err := g.ResolveChooseCards(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("choose none: %v", err)
	}
	effectstest.AnswerScryIfAsked(t, g, me)

	if _, ok := effectstest.BattlefieldCardByID(g, fortune); ok {
		t.Error("Fortune was not flickered")
	}
	if ids := effectstest.BattlefieldIDsNamed(g, "Fortune, Loyal Steed"); len(ids) != 1 {
		t.Errorf("%d Fortunes on the battlefield, want 1", len(ids))
	}
	if _, ok := effectstest.BattlefieldCardByID(g, saddler); !ok {
		t.Error("the saddler was flickered though none was chosen")
	}
}

// The ruling: the delayed trigger works whether or not Fortune is on the
// battlefield, so its controller can still exile a creature that saddled it.
func TestFortuneKilledBeforeEndOfCombatStillFlickersASaddler(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := effectstest.PushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", effectstest.FortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	saddler := effectstest.PushCrewerForTest(g, me.ID, "Saddler", 1)

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
	if _, ok := effectstest.BattlefieldCardByID(g, saddler); ok {
		t.Error("the saddler was not flickered")
	}
	if ids := effectstest.BattlefieldIDsNamed(g, "Saddler"); len(ids) != 1 {
		t.Errorf("%d Saddlers on the battlefield, want 1 (returned)", len(ids))
	}
	if !me.Graveyard.Contains(fortune) {
		t.Error("the dead Fortune left the graveyard")
	}
}

// A token that saddled it is exiled and does not come back (CR 111.8).
func TestFortuneTokenSaddlerDoesNotReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortune := effectstest.PushDiesCreatureForTest(g, me.ID, "Fortune, Loyal Steed", effectstest.FortuneLoyalSteedOracle, "Legendary Creature — Beast Mount", 2, 4)
	token := effectstest.PushDiesCreatureForTest(g, me.ID, "Test Token", "", "Token Creature — Test", 1, 1)

	fortuneAttacks(t, g, me, opp, fortune, token)
	choice := fortuneEndOfCombatPrompt(t, g, me)
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{token}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.AnswerScryIfAsked(t, g, me)

	if ids := effectstest.BattlefieldIDsNamed(g, "Test Token"); len(ids) != 0 {
		t.Errorf("the token came back (%d on the battlefield)", len(ids))
	}
	if ids := effectstest.BattlefieldIDsNamed(g, "Fortune, Loyal Steed"); len(ids) != 1 {
		t.Errorf("%d Fortunes on the battlefield, want 1", len(ids))
	}
}
