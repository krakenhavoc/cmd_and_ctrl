package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seam_5425_cards_test.go — #5425: the explore replacement "explores, then
// explores again" (Topography Tracker), several permanents exploring at
// once (CR 701.44d, Hakbal of the Surging Soul) and investigate as a
// keyword action with its own event (CR 701.16a, Erdwal Illuminator).

const (
	seam5425TrackerOracle = topographyTrackerOracleID
	seam5425HakbalOracle  = "687e772a-cdfe-41fd-a066-6cb3aee04193"
	seam5425ErdwalOracle  = "5ff4ef34-acb9-46c3-8530-f4b093807066"
	seam5425ThrabenOracle = "caa02547-66e3-4e27-a2d3-5e94f3e7a069"
)

func TestSeam5425CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		seam5425TrackerOracle: "Topography Tracker",
		seam5425HakbalOracle:  "Hakbal of the Surging Soul",
		seam5425ErdwalOracle:  "Erdwal Illuminator",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func seam5425Tracker(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b21Push(g, owner, "Topography Tracker", "Creature — Merfolk Scout", seam5425TrackerOracle, 2, 2, "G")
}

func seam5425Explored(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventExplored && ev.CardID == id {
			n++
		}
	}
	return n
}

func seam5425Count(g *game.Game, kind game.EventKind, actor uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.Actor == actor {
			n++
		}
	}
	return n
}

func seam5425Named(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

func seam5425Explore(g *game.Game, id, controller uuid.UUID) {
	g.WithWriteLock(func() {
		ref := game.ObjectRef{ID: id}
		if c, ok := g.LookupCardForEffect(id); ok {
			ref.Epoch = c.ObjectEpoch
		}
		_ = g.ExploreForEffect(uuid.Nil, ref, controller)
	})
}

// --- Topography Tracker ------------------------------------------------------

func TestSeam5425TopographyTrackerMakesAMapAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{Name: "Topography Tracker", TypeLine: "Creature — Merfolk Scout", OracleID: seam5425TrackerOracle, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Map"); got != 1 {
		t.Errorf("controller has %d Map tokens, want 1", got)
	}
}

// The second explore starts once the first is complete (CR 701.44b), so a
// nonland card left on top is revealed again, and each is an explore.
func TestSeam5425TopographyTrackerExploresThenExploresAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Tracker(g, me.ID)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
	advanceToMain(t, g)

	seam5425Explore(g, bear, me.ID)
	if got := plusOneCounters(g, bear); got != 1 {
		t.Fatalf("after the first reveal the explorer has %d +1/+1 counters, want 1", got)
	}
	if got := seam5425Explored(g, bear); got != 0 {
		t.Fatalf("explored %d times before the first question was answered", got)
	}
	answerExploreQuestion(t, g, me.ID, false)
	if got := seam5425Explored(g, bear); got != 1 {
		t.Fatalf("explored %d times after the first answer, want 1", got)
	}
	if got := plusOneCounters(g, bear); got != 2 {
		t.Fatalf("the card left on top was not revealed again: %d +1/+1 counters, want 2", got)
	}
	answerExploreQuestion(t, g, me.ID, true)
	if !me.Graveyard.Contains(top) {
		t.Error("the second explore did not put the card into the graveyard")
	}
	if got := seam5425Explored(g, bear); got != 2 {
		t.Errorf("explored %d times, want 2", got)
	}
	if latestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("a third explore asked its question")
	}
}

func TestSeam5425TopographyTrackerALandThenANonland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Tracker(g, me.ID)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	putOnTopForTest(me, "Some Sorcery", "Sorcery")
	land := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)

	seam5425Explore(g, bear, me.ID)
	if !me.Hand.Contains(land) {
		t.Fatal("the first explore did not put the land into hand")
	}
	if got := plusOneCounters(g, bear); got != 1 {
		t.Fatalf("the second explore gave %d +1/+1 counters, want 1", got)
	}
	answerExploreQuestion(t, g, me.ID, false)
	if got := seam5425Explored(g, bear); got != 2 {
		t.Errorf("explored %d times, want 2", got)
	}
}

// Each Tracker applies once to what the one before it made (CR 614.5).
func TestSeam5425TwoTopographyTrackersAreFourExplores(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Tracker(g, me.ID)
	seam5425Tracker(g, me.ID)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	putOnTopForTest(me, "Some Sorcery", "Sorcery")
	advanceToMain(t, g)

	seam5425Explore(g, bear, me.ID)
	for i := 0; i < 4; i++ {
		answerExploreQuestion(t, g, me.ID, false)
	}
	if got := seam5425Explored(g, bear); got != 4 {
		t.Errorf("explored %d times, want 4", got)
	}
	if got := plusOneCounters(g, bear); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
	if latestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("a fifth explore asked its question")
	}
}

// "A creature you control": an opponent's creature explores once.
func TestSeam5425TopographyTrackerLeavesAnOpponentsExploreAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seam5425Tracker(g, me.ID)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	putOnTopForTest(opp, "Some Sorcery", "Sorcery")
	advanceToMain(t, g)

	seam5425Explore(g, theirs, opp.ID)
	answerExploreQuestion(t, g, opp.ID, false)
	if got := seam5425Explored(g, theirs); got != 1 {
		t.Errorf("explored %d times, want 1", got)
	}
	if latestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("the opponent's creature explored again")
	}
}

// With Twists and Turns the event is one scry and then both explores, in
// either order of applying the two replacements (CR 616.1, 614.5).
func TestSeam5425TopographyTrackerWithTwistsAndTurnsScriesOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Tracker(g, me.ID)
	b21Push(g, me.ID, "Twists and Turns", "Enchantment", twistsAndTurnsOracleID, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	putOnTopForTest(me, "Some Sorcery", "Sorcery")
	advanceToMain(t, g)

	seam5425Explore(g, bear, me.ID)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry before the explores")
	}
	if got := plusOneCounters(g, bear); got != 0 {
		t.Fatal("the creature explored before the scry was answered")
	}
	answerScryKeepAll(t, g, me.ID)
	answerExploreQuestion(t, g, me.ID, false)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("the explore the Tracker added was replaced by a second scry")
	}
	answerExploreQuestion(t, g, me.ID, false)
	if got := seam5425Explored(g, bear); got != 2 {
		t.Errorf("explored %d times, want 2", got)
	}
	if got := seam5425Count(g, game.EventScry, me.ID); got != 1 {
		t.Errorf("scried %d times, want 1", got)
	}
}

// The Map it makes, used on the Tracker itself: two explores.
func TestSeam5425AMapUnderATopographyTrackerExploresTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tracker := seam5425Tracker(g, me.ID)
	m := pushToken(g, me.ID, MapToken())
	putOnTopForTest(me, "Island", "Basic Land — Island")
	putOnTopForTest(me, "Forest", "Basic Land — Forest")
	hand := me.Hand.Size()

	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tracker}},
	}); err != nil {
		t.Fatalf("activate the Map: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2 (two lands explored)", got)
	}
	if got := seam5425Explored(g, tracker); got != 2 {
		t.Errorf("explored %d times, want 2", got)
	}
}

// --- Hakbal of the Surging Soul ---------------------------------------------

func seam5425Hakbal(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b21Push(g, owner, "Hakbal of the Surging Soul", "Legendary Creature — Merfolk Scout", seam5425HakbalOracle, 3, 3, "G", "U")
}

// CR 701.44d: you choose which Merfolk explores, it explores, and you
// choose again knowing what the first one left on top.
func TestSeam5425HakbalEachMerfolkExploresInAnOrderYouChoose(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hakbal := seam5425Hakbal(g, me.ID)
	fish := b16Creature(g, me.ID, "Fish", "Creature — Merfolk Wizard", 1, 1, "U")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")

	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no choice of which Merfolk explores first")
	}
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, hakbal) || !hasID(pick.ChooseCards, fish) || hasID(pick.ChooseCards, bear) {
		t.Fatalf("offered %v, want Hakbal and the other Merfolk only", pick.ChooseCards)
	}
	if plusOnes(g, hakbal) != 0 || plusOnes(g, fish) != 0 {
		t.Fatal("a Merfolk explored before the choice was made")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{fish}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if plusOnes(g, fish) != 1 || plusOnes(g, hakbal) != 0 {
		t.Fatalf("counters after the first explore: fish %d, Hakbal %d; want 1, 0", plusOnes(g, fish), plusOnes(g, hakbal))
	}
	// Left on top: the last Merfolk reveals it again, with no choice to make.
	answerExploreQuestion(t, g, me.ID, false)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Fatal("asked which explores next with one Merfolk left")
	}
	if plusOnes(g, hakbal) != 1 {
		t.Fatalf("Hakbal has %d +1/+1 counters, want 1", plusOnes(g, hakbal))
	}
	answerExploreQuestion(t, g, me.ID, true)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed card was not put into the graveyard")
	}
	if seam5425Explored(g, fish) != 1 || seam5425Explored(g, hakbal) != 1 || seam5425Explored(g, bear) != 0 {
		t.Errorf("explored: fish %d, Hakbal %d, bear %d; want 1, 1, 0", seam5425Explored(g, fish), seam5425Explored(g, hakbal), seam5425Explored(g, bear))
	}
	if plusOnes(g, bear) != 0 {
		t.Error("a creature that is not a Merfolk explored")
	}
}

func TestSeam5425HakbalAloneExploresWithoutAChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hakbal := seam5425Hakbal(g, me.ID)
	land := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Fatal("asked which explores with one Merfolk")
	}
	if !me.Hand.Contains(land) || seam5425Explored(g, hakbal) != 1 {
		t.Errorf("Hakbal explored %d times, land in hand %v", seam5425Explored(g, hakbal), me.Hand.Contains(land))
	}
}

func TestSeam5425HakbalDoesNotTriggerOnAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	hakbal := seam5425Hakbal(g, opp.ID)
	putOnTopForTest(opp, "Forest", "Basic Land — Forest")
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if seam5425Explored(g, hakbal) != 0 {
		t.Error("Hakbal explored at the beginning of another player's combat")
	}
}

// With a Topography Tracker each Merfolk's explore is two, and the Tracker
// is a Merfolk itself.
func TestSeam5425HakbalWithATopographyTracker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hakbal := seam5425Hakbal(g, me.ID)
	tracker := seam5425Tracker(g, me.ID)
	for i := 0; i < 4; i++ {
		putOnTopForTest(me, "Forest", "Basic Land — Forest")
	}
	hand := me.Hand.Size()
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no choice of which Merfolk explores first")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{tracker}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if seam5425Explored(g, hakbal) != 2 || seam5425Explored(g, tracker) != 2 {
		t.Errorf("explored: Hakbal %d, Tracker %d; want 2 each", seam5425Explored(g, hakbal), seam5425Explored(g, tracker))
	}
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("hand grew by %d, want 4", got)
	}
}

func seam5425HakbalAttacks(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	hakbal := seam5425Hakbal(g, me.ID)
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	// Hakbal explored at the beginning of combat; leave a nonland on top.
	if latestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		answerExploreQuestion(t, g, me.ID, false)
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(hakbal, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	return hakbal
}

func TestSeam5425HakbalAttacksAndPutsALandFromHandOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := mshA6HandCard(me, "Island", "Basic Land — Island", "")
	spell := mshA6HandCard(me, "Some Sorcery", "Sorcery", "{1}{U}")
	seam5425HakbalAttacks(t, g, me, opp)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the attack trigger offered nothing")
	}
	if !hasID(pick.ChooseCards, land) || hasID(pick.ChooseCards, spell) {
		t.Errorf("candidates %v: want only the land", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not enter")
	}
	if got := me.Hand.Size(); got != hand-1 {
		t.Errorf("hand %d, want %d: no draw when a land was put", got, hand-1)
	}
}

func TestSeam5425HakbalDrawsWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := mshA6HandCard(me, "Island", "Basic Land — Island", "")
	seam5425HakbalAttacks(t, g, me, opp)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(land) {
		t.Error("a declined land entered")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: declining draws a card", got, hand+1)
	}
}

func TestSeam5425HakbalDrawsWithNoLandInHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for me.Hand.Size() > 0 {
		c, _ := me.Hand.PopTop()
		me.Library.PushBottom(c)
	}
	mshA6HandCard(me, "Some Sorcery", "Sorcery", "{1}{U}")
	seam5425HakbalAttacks(t, g, me, opp)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a prompt opened with no land in hand")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: no land to put draws a card", got, hand+1)
	}
}

// --- Erdwal Illuminator ------------------------------------------------------

func seam5425Erdwal(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b21Push(g, owner, "Erdwal Illuminator", "Creature — Spirit", seam5425ErdwalOracle, 1, 3, "U")
}

func seam5425Investigate(g *game.Game, player uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.InvestigateForEffect(player, uuid.Nil, n) })
}

func TestSeam5425ErdwalIlluminatorInvestigatesAnAdditionalTimeOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Erdwal(g, me.ID)
	advanceToMain(t, g)

	seam5425Investigate(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Clue"); got != 2 {
		t.Fatalf("%d Clues after the turn's first investigate, want 2", got)
	}
	if got := seam5425Count(g, game.EventInvestigate, me.ID); got != 2 {
		t.Errorf("investigated %d times, want 2", got)
	}
	// The additional investigate was the turn's second, and so is not a first.
	seam5425Investigate(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Clue"); got != 3 {
		t.Errorf("%d Clues after a later investigate the same turn, want 3", got)
	}

	// A new turn has a new first.
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	seam5425Investigate(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Clue"); got != 5 {
		t.Errorf("%d Clues after the next turn's first investigate, want 5", got)
	}
}

// "Investigate twice" is two investigates, only one of them the first.
func TestSeam5425ErdwalIlluminatorOnInvestigateTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Erdwal(g, me.ID)
	advanceToMain(t, g)
	seam5425Investigate(g, me.ID, 2)
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Clue"); got != 3 {
		t.Errorf("%d Clues, want 3", got)
	}
}

func TestSeam5425TwoErdwalIlluminatorsEachTriggerOnTheFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Erdwal(g, me.ID)
	seam5425Erdwal(g, me.ID)
	advanceToMain(t, g)
	seam5425Investigate(g, me.ID, 1)
	for i := 0; i < 4 && !stackFullyEmpty(g); i++ {
		if latestChoiceOfKindFor(g, game.PendingChoiceTriggerOrder, me.ID) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
		}
		passPriorityAroundTable(t, g)
	}
	if got := seam5425Named(g, me.ID, "Clue"); got != 3 {
		t.Errorf("%d Clues, want 3", got)
	}
}

// A Clue made without the word is not an investigate, and an opponent's
// investigate is not yours.
func TestSeam5425ErdwalIlluminatorIgnoresAPlainClueAndAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seam5425Erdwal(g, me.ID)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, ClueToken(), 1) })
	seam5425Investigate(g, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Clue"); got != 1 {
		t.Errorf("I have %d Clues, want the 1 I created", got)
	}
	if got := seam5425Named(g, opp.ID, "Clue"); got != 1 {
		t.Errorf("the opponent has %d Clues, want 1", got)
	}
	// My own first investigate of the turn still is one.
	seam5425Investigate(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Clue"); got != 3 {
		t.Errorf("I have %d Clues after investigating, want 3", got)
	}
}

// An Illuminator that enters after the turn's first investigate saw none.
func TestSeam5425ErdwalIlluminatorEnteringAfterTheFirstInvestigate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seam5425Investigate(g, me.ID, 1)
	seam5425Erdwal(g, me.ID)
	seam5425Investigate(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := seam5425Named(g, me.ID, "Clue"); got != 2 {
		t.Errorf("%d Clues, want 2", got)
	}
}

// End to end through a printed card: Thraben Inspector's "investigate".
func TestSeam5425ErdwalIlluminatorSeesAPrintedInvestigate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seam5425Erdwal(g, me.ID)
	msc17Play(t, g, me, game.Card{Name: "Thraben Inspector", TypeLine: "Creature — Human Soldier", OracleID: seam5425ThrabenOracle, Power: 1, Toughness: 2})
	for i := 0; i < 3; i++ {
		passPriorityAroundTable(t, g)
	}
	if got := seam5425Named(g, me.ID, "Clue"); got != 2 {
		t.Errorf("%d Clues, want 2", got)
	}
}
