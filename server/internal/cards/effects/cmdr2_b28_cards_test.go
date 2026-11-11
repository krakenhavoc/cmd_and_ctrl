package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2B28SpybugOracle    = "3bb7041c-4f3e-4123-9181-3442ea0c98e8"
	cmdr2B28PhantasmOracle  = "ca14f1b6-e8c2-4ea5-9375-5ed63f355052"
	cmdr2B28SnitchOracle    = "b6836023-4475-415e-9671-980d9eb6e762"
	cmdr2B28CatchersOracle  = "d8e276a8-1234-4b61-84cf-4bf231ea3129"
	cmdr2B28EnhancedOracle  = "bfb5104b-01e3-4d37-ac93-1dc342121531"
	cmdr2B28CampaignOracle  = "1dd336b0-9b13-4180-b80a-e8ba5d63d04f"
	cmdr2B28DreamEaterOracl = "4ec84bb3-796a-4074-ac9a-1bd277158896"
	cmdr2B28DetectiveOracle = "c5a1516a-5e01-48c5-88c7-2178f4b2ad24"
	cmdr2B28MirkoOracle     = "1692f19e-4906-4808-965f-5f0284004f55"
)

// cmdr2B28KeepSurveil answers the open surveil prompt by keeping every card
// on top.
func cmdr2B28KeepSurveil(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

// cmdr2B28ConsiderSurveil surveils 1 for `me` by casting Consider and
// answering the prompt (keep it on top), then lets the triggers resolve.
func cmdr2B28ConsiderSurveil(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	castCatalogSpell(t, g, "Consider", "Instant", considerOracle, nil)
	passPriorityAroundTable(t, g)
	cmdr2B28KeepSurveil(t, g, me)
}

func cmdr2B28Library(me *game.Player) {
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
}

func TestCmdr2B28DimirSpybugGrowsWhenYouSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B28Library(me)
	bug := b12Push(g, me.ID, "Dimir Spybug", "Creature — Insect", cmdr2B28SpybugOracle, 1, 1)
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bug, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d, want 1", n)
	}
	for _, kw := range []string{"flying", "menace"} {
		if !p1g2HasKeyword(g, bug, kw) {
			t.Errorf("Dimir Spybug lacks %s", kw)
		}
	}
}

func TestCmdr2B28DimirSpybugIgnoresAnOpponentsSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B28Library(opp)
	bug := b12Push(g, opp.ID, "Dimir Spybug", "Creature — Insect", cmdr2B28SpybugOracle, 1, 1)
	cmdr2B28Library(me)
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bug, game.CounterPlusOne); n != 0 {
		t.Errorf("counters %d from my surveil, want 0", n)
	}
}

func TestCmdr2B28ThoughtboundPhantasmGrowsAndLosesItsDefenderAtThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	cmdr2B28Library(me)
	ph := b12Push(g, me.ID, "Thoughtbound Phantasm", "Creature — Spirit", cmdr2B28PhantasmOracle, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ph, opp.ID); err == nil {
		t.Fatal("a defender with no counters attacked")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ph, game.CounterPlusOne, 2) })
	if err := g.DeclareAttacker(ph, opp.ID); err == nil {
		t.Fatal("a defender with two counters attacked")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ph, game.CounterPlusOne, 1) })
	if err := g.DeclareAttacker(ph, opp.ID); err != nil {
		t.Fatalf("a Phantasm with three counters could not attack: %v", err)
	}
}

func TestCmdr2B28ThoughtboundPhantasmGetsACounterPerSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B28Library(me)
	ph := b12Push(g, me.ID, "Thoughtbound Phantasm", "Creature — Spirit", cmdr2B28PhantasmOracle, 2, 2)
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ph, game.CounterPlusOne); n != 1 {
		t.Errorf("counters %d, want 1", n)
	}
}

func TestCmdr2B28WhisperingSnitchTriggersOnlyOnTheFirstSurveilEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B28Library(me)
	b12Push(g, me.ID, "Whispering Snitch", "Creature — Vampire Rogue", cmdr2B28SnitchOracle, 1, 3)
	life, oppLife := me.Life, opp.Life
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Fatalf("life %d/%d, want %d/%d after the first surveil", me.Life, opp.Life, life+1, oppLife-1)
	}
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("the second surveil of the turn triggered it again: %d/%d", me.Life, opp.Life)
	}
}

func TestCmdr2B28DisinformationCampaignDrawsDiscardsAndBouncesOnSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B28Library(me)
	hand, oppHand := me.Hand.Size(), opp.Hand.Size()
	castCatalogSpell(t, g, "Disinformation Campaign", "Enchantment", cmdr2B28CampaignOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 { // the Campaign left the hand, one card drawn
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	for _, o := range g.Seats[1:] { // every opponent discards
		c := discardChoiceFor(g, o.ID)
		if c == nil {
			t.Fatalf("no discard prompt for %s", o.Name)
		}
		if err := g.ResolveChooseCards(c.ID, o.ID, c.ChooseCards[:1]); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if opp.Hand.Size() != oppHand-1 {
		t.Errorf("opponent hand %d, want %d", opp.Hand.Size(), oppHand-1)
	}
	id := findBattlefieldByName(g, "Disinformation Campaign")
	if id == uuid.Nil {
		t.Fatal("Disinformation Campaign is not on the battlefield")
	}
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
		t.Error("Disinformation Campaign did not return to its owner's hand")
	}
}

func TestCmdr2B28CopyCatchersPaysForAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B28Library(me)
	cc := b12Push(g, me.ID, "Copy Catchers", "Creature — Faerie", cmdr2B28CatchersOracle, 2, 1)
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "U", 2)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Copy Catchers" && c.InstanceID != cc && c.IsToken() {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d token copies, want 1", n)
	}
}

func TestCmdr2B28CopyCatchersDeclineMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B28Library(me)
	b12Push(g, me.ID, "Copy Catchers", "Creature — Faerie", cmdr2B28CatchersOracle, 2, 1)
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "U", 2)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Copy Catchers" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Copy Catchers, want just the original", n)
	}
}

func TestCmdr2B28EnhancedSurveillanceLooksAtTwoMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B28Library(me)
	b12Permanent(g, me.ID, "Enhanced Surveillance", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Enhanced Surveillance" {
				g.Battlefield.Cards[i].OracleID = cmdr2B28EnhancedOracle
			}
		}
	})
	castCatalogSpell(t, g, "Consider", "Instant", considerOracle, nil)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceOptionalReplacement)
	if offer == nil {
		t.Fatalf("no optional replacement offer: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 3 {
		t.Fatalf("want a surveil over three cards, got %+v", c)
	}
}

func TestCmdr2B28EnhancedSurveillanceCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B28Library(me)
	id := b12Permanent(g, me.ID, "Enhanced Surveillance", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].OracleID = cmdr2B28EnhancedOracle
			}
		}
	})
	castCatalogSpell(t, g, "Consider", "Instant", considerOracle, nil)
	passPriorityAroundTable(t, g)
	if offer := latestChoiceOfKind(g, game.PendingChoiceOptionalReplacement); offer != nil {
		if err := g.ResolveOptionalReplacement(offer.ID, me.ID, false); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a plain surveil 1, got %+v", c)
	}
}

func TestCmdr2B28EnhancedSurveillanceExilesToShuffleYourGraveyardIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Permanent(g, me.ID, "Enhanced Surveillance", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].OracleID = cmdr2B28EnhancedOracle
			}
		}
	})
	pushGraveyardCardForTest(me, "One")
	pushGraveyardCardForTest(me, "Two")
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(id) || !g.Exile.Contains(id) {
		t.Error("the enchantment should be exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 0 || me.Library.Size() != lib+2 {
		t.Errorf("graveyard %d, library %d (was %d), want an empty graveyard and two more cards", me.Graveyard.Size(), me.Library.Size(), lib)
	}
}

func TestCmdr2B28DreamEaterSurveilsFourThenBouncesAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B28Library(me)
	theirs := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Ox", 2, 2)
	land := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Dream Eater", "Creature — Nightmare Sphinx", cmdr2B28DreamEaterOracl, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 4 {
		t.Fatalf("want a surveil 4 prompt, got %+v", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	// The reflexive trigger targets a nonland permanent an opponent controls.
	b04WaitForPick(t, g, me.ID)
	pk := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pk.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("my own creature was a legal target")
	}
	if err := g.ResolvePickTarget(pk.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Error("a land was a legal target")
	}
	pickCard(t, g, me.ID, theirs)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the Ox was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature was bounced")
	}
}

func TestCmdr2B28DoggedDetectiveSurveilsTwoWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr2B28Library(me)
	castCatalogSpell(t, g, "Dogged Detective", "Creature — Human Rogue", cmdr2B28DetectiveOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil 2 prompt, got %+v", c)
	}
}

func TestCmdr2B28DoggedDetectiveReturnsWhenAnOpponentDrawsTheirSecondCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dogged Detective", TypeLine: "Creature — Human Rogue",
		OracleID: cmdr2B28DetectiveOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if hasPromptFor(g, me.ID) {
		t.Fatal("the first draw of the turn triggered it")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("Dogged Detective did not return to hand")
	}
}

func TestCmdr2B28DoggedDetectiveIgnoresYourOwnSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dogged Detective", TypeLine: "Creature — Human Rogue",
		OracleID: cmdr2B28DetectiveOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	if hasPromptFor(g, me.ID) || !me.Graveyard.Contains(id) {
		t.Error("my own draws should not return it")
	}
}

func hasPromptFor(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == chooser {
			return true
		}
	}
	return false
}

func TestCmdr2B28MirkoGrowsAndReturnsACreatureWithAFinalityCounterAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cmdr2B28Library(me)
	mirko := b12Push(g, me.ID, "Mirko, Obsessive Theorist", "Legendary Creature — Vampire Detective", cmdr2B28MirkoOracle, 1, 3)
	weak := pushGraveyardCardForTest(me, "Weak") // 1/1: power 1 is not less than Mirko's 1
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == weak {
				me.Graveyard.Cards[i].Power = 1
			}
		}
	})
	cmdr2B28ConsiderSurveil(t, g, me)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, mirko, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d, want 1", n)
	}
	// Mirko is now 2/4: the 1-power creature card is a legal target.
	advanceTo(t, g, game.StepEnd)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, weak)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, weak)
	if !ok {
		t.Fatal("the creature card did not return")
	}
	if c.Counters[game.CounterFinality] != 1 {
		t.Errorf("finality counters %d, want 1", c.Counters[game.CounterFinality])
	}
}

func TestCmdr2B28MirkoCannotReturnAnEqualOrGreaterPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Mirko, Obsessive Theorist", "Legendary Creature — Vampire Detective", cmdr2B28MirkoOracle, 1, 3)
	big := pushGraveyardCardForTest(me, "Big")
	advanceTo(t, g, game.StepEnd)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("offered a target with no creature card of lesser power")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the creature card should stay in the graveyard")
	}
}
