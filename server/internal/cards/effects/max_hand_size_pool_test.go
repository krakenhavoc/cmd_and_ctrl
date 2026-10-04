package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The #2074 pool (ADR 0113 §3, owner decision 2): every catalog card
// the maximum-hand-size seam alone unblocks. The timestamp-order tests
// follow each card's own ruling.

const (
	mhGnatMiserOracle       = "cfb5da72-9f00-4ae2-8d59-f997faab451b"
	mhLocustMiserOracle     = "75a53a88-51a5-4c76-9f67-ee94c3065b98"
	mhThoughtNibblerOracle  = "9a640695-fc17-4ccc-ac88-ce48b6c6ad53"
	mhThoughtEaterOracle    = "88e18dc9-8b11-4369-9860-a687925cdcb4"
	mhThoughtDevourerOracle = "e5106a2d-caca-4519-a1a1-eabaec4005df"
	mhMinamoOracle          = "9ad4f876-e923-4a09-8e70-eb826aacf89d"
	mhRecycleOracle         = "8cef3ce7-8fbc-4e68-93f9-d6adfc2c2cbf"
	mhCursedRackOracle      = "4f04603f-8f91-405b-b6ac-d2b66f05e32f"
	mhTrustedAdvisorOracle  = "430e4e74-33ee-49f2-aef5-6d09079c5eb7"
	mhDoctorOctopusOracle   = "2fecfbef-e521-4025-94ea-451c9abde3de"
	mhTenRingsOracle        = "a82d855c-50c8-43f2-99ba-a84fe84539c5"
	mhToadOracle            = "cce643fb-cd88-457c-8f13-c203554df675"
	mhAnvilOracle           = "9deb2a0b-f40c-4d13-8321-41dc41448d24"
	mhFolioOracle           = "9030d508-5cfe-46ed-954b-21d970bc3b5a"
	mhMidnightOilOracle     = "26bfc19c-8122-4e95-9a62-2b5418429db2"
	mhWinterOracle          = "5a9baafb-bffd-4e28-bbe1-b4154cd86bf3"
)

// mhResolveAll passes priority until the stack is empty, ordering any
// simultaneous triggers as they were queued.
func mhResolveAll(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		answerTriggerOrderLastQueuedFirst(t, g)
		passPriorityAroundTable(t, g)
		if triggerOrderPrompt(g) == nil {
			return
		}
	}
}

// mhRack puts a Cursed Rack under `controller` with `chosen` already
// chosen (the as-enters choice is ChoosePlayerAsEnters's, tested with
// True-Name Nemesis).
func mhRack(g *game.Game, controller, chosen uuid.UUID) uuid.UUID {
	id := mhPush(g, controller, "Cursed Rack", "Artifact", mhCursedRackOracle)
	findBattlefieldCardForTest(g, id).ChosenPlayer = chosen
	return id
}

// The Gnat Miser ruling (2009-10-01): Cursed Rack on the opponent, then
// Gnat Miser, is three; the other order is four.
func TestGnatMiserAndCursedRackFollowTimestampOrder(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhRack(g, me.ID, opp.ID)
	mhPush(g, me.ID, "Gnat Miser", "Creature — Rat Shaman", mhGnatMiserOracle)
	if got := mhMax(g, opp); got != 3 {
		t.Errorf("Cursed Rack, then Gnat Miser: %d, want 3", got)
	}
	if got := mhMax(g, me); got != 7 {
		t.Errorf("Gnat Miser's controller: %d, want 7", got)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	mhPush(g, me.ID, "Gnat Miser", "Creature — Rat Shaman", mhGnatMiserOracle)
	mhRack(g, me.ID, opp.ID)
	if got := mhMax(g, opp); got != 4 {
		t.Errorf("Gnat Miser, then Cursed Rack: %d, want 4", got)
	}
}

// Cursed Rack reaches the chosen player only, and nobody until one is
// chosen.
func TestCursedRackReachesTheChosenPlayerOnly(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	rack := mhPush(g, me.ID, "Cursed Rack", "Artifact", mhCursedRackOracle)
	if got := mhMax(g, opp); got != 7 {
		t.Errorf("nobody chosen yet: %d, want 7", got)
	}
	findBattlefieldCardForTest(g, rack).ChosenPlayer = opp.ID
	for p, want := range map[*game.Player]int{me: 7, opp: 4, other: 7} {
		if got := mhMax(g, p); got != want {
			t.Errorf("%s: %d, want %d", p.Name, got, want)
		}
	}
}

func TestLocustMiserReducesEachOpponentByTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mhPush(g, me.ID, "Locust Miser", "Creature — Rat Shaman", mhLocustMiserOracle)
	for _, p := range g.Seats {
		want := 5
		if p.ID == me.ID {
			want = 7
		}
		if got := mhMax(g, p); got != want {
			t.Errorf("%s: %d, want %d", p.Name, got, want)
		}
	}
}

// The Thought Beasts reduce their controller's own maximum, and fly.
func TestThoughtBeastsReduceTheirController(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		want         int
	}{
		{"Thought Nibbler", mhThoughtNibblerOracle, 5},
		{"Thought Eater", mhThoughtEaterOracle, 4},
		{"Thought Devourer", mhThoughtDevourerOracle, 3},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := mhPush(g, me.ID, tc.name, "Creature — Beast", tc.oracle)
		if got := mhMax(g, me); got != tc.want {
			t.Errorf("%s: controller's maximum %d, want %d", tc.name, got, tc.want)
		}
		if got := mhMax(g, opp); got != 7 {
			t.Errorf("%s: an opponent's maximum %d, want 7", tc.name, got)
		}
		if !game.HasKeyword(findBattlefieldCardForTest(g, id), "flying") {
			t.Errorf("%s has no flying", tc.name)
		}
	}
}

// The Thought Eater ruling (2009-10-01): under a Cursed Rack, then
// Thought Eater, the maximum is one; the other order is four.
func TestThoughtEaterAndCursedRackFollowTimestampOrder(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhRack(g, opp.ID, me.ID)
	mhPush(g, me.ID, "Thought Eater", "Creature — Beast", mhThoughtEaterOracle)
	if got := mhMax(g, me); got != 1 {
		t.Errorf("Cursed Rack, then Thought Eater: %d, want 1", got)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	mhPush(g, me.ID, "Thought Eater", "Creature — Beast", mhThoughtEaterOracle)
	mhRack(g, opp.ID, me.ID)
	if got := mhMax(g, me); got != 4 {
		t.Errorf("Thought Eater, then Cursed Rack: %d, want 4", got)
	}
}

// The Minamo Scrollkeeper ruling (2013-04-15): Null Profusion then
// Scrollkeeper is three; the other order is two.
func TestMinamoScrollkeeperAndNullProfusionFollowTimestampOrder(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[0]
	mhNullProfusion(g, me.ID)
	mhPush(g, me.ID, "Minamo Scrollkeeper", "Creature — Human Wizard", mhMinamoOracle)
	if got := mhMax(g, me); got != 3 {
		t.Errorf("Null Profusion, then Scrollkeeper: %d, want 3", got)
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	mhPush(g, me.ID, "Minamo Scrollkeeper", "Creature — Human Wizard", mhMinamoOracle)
	mhNullProfusion(g, me.ID)
	if got := mhMax(g, me); got != 2 {
		t.Errorf("Scrollkeeper, then Null Profusion: %d, want 2", got)
	}
}

// Recycle is Null Profusion's twin: the Spellbook ruling in both orders,
// and a land played draws a card.
func TestRecycle(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mhSpellbook(g, me.ID)
	mhPush(g, me.ID, "Recycle", "Enchantment", mhRecycleOracle)
	if got := mhMax(g, me); got != 2 {
		t.Errorf("Spellbook, then Recycle: %d, want 2", got)
	}
	advanceToMain(t, g)
	library := me.Library.Size()
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Forest: %v", err)
	}
	passPriorityAroundTable(t, g)
	if drawn := library - me.Library.Size(); drawn != 1 {
		t.Errorf("playing a land drew %d, want 1", drawn)
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	mhPush(g, me.ID, "Recycle", "Enchantment", mhRecycleOracle)
	mhSpellbook(g, me.ID)
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("Recycle, then Spellbook: %d, want no maximum", got)
	}
}

// The Trusted Advisor ruling (2009-10-01): Null Profusion then Trusted
// Advisor is four. Its upkeep returns a blue creature you control, the
// Advisor itself included.
func TestTrustedAdvisor(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[0]
	mhNullProfusion(g, me.ID)
	advisor := mhPush(g, me.ID, "Trusted Advisor", "Creature — Human Advisor", mhTrustedAdvisorOracle)
	findBattlefieldCardForTest(g, advisor).ManaCost = "{U}"
	if got := mhMax(g, me); got != 4 {
		t.Errorf("Null Profusion, then Trusted Advisor: %d, want 4", got)
	}

	green := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	findBattlefieldCardForTest(g, green).ManaCost = "{1}{G}"
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	advanceToStepOf(t, g, 0, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Trusted Advisor queued no return pick")
	}
	if !hasID(p.ChooseCards, advisor) || hasID(p.ChooseCards, green) {
		t.Errorf("offered %v, want the blue Advisor and not the green Bear", p.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{advisor}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !me.Hand.Contains(advisor) {
		t.Error("the Advisor should be back in its owner's hand")
	}
}

// Doctor Octopus: other Villains you control get +2/+2, the maximum is
// eight, and the end step draws up to eight.
func TestDoctorOctopusMasterPlanner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	doc := mhPush(g, me.ID, "Doctor Octopus, Master Planner", "Legendary Creature — Human Scientist Villain", mhDoctorOctopusOracle)
	villain := b12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	theirs := b12Creature(g, opp.ID, "Rival", "Creature — Human Villain", 1, 1)
	if got := mhMax(g, me); got != 8 {
		t.Errorf("maximum %d, want 8", got)
	}
	var vp, tp, dp int
	g.ReadSnapshot(func() {
		vp = findBattlefieldCardForTest(g, villain).Effective().Power
		tp = findBattlefieldCardForTest(g, theirs).Effective().Power
		dp = findBattlefieldCardForTest(g, doc).Effective().Power
	})
	if vp != 3 || tp != 1 {
		t.Errorf("your Villain %d, their Villain %d; want 3 and 1", vp, tp)
	}
	if dp != 1 {
		t.Errorf("Doctor Octopus pumps itself: power %d, want its own 1", dp)
	}

	for me.Hand.Size() > 3 {
		me.Hand.Cards = me.Hand.Cards[1:]
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 8 {
		t.Errorf("hand after the end step %d, want 8", got)
	}
}

// The Ten Rings: maximum ten, and the end step tops the hand up to ten,
// but only when it is short.
func TestTheTenRings(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mhPush(g, me.ID, "The Ten Rings", "Legendary Artifact", mhTenRingsOracle)
	if got := mhMax(g, me); got != 10 {
		t.Errorf("maximum %d, want 10", got)
	}
	fillHandTo(t, g, me, 4)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 10 || hand >= 10 {
		t.Errorf("hand %d -> %d after the end step, want 10", hand, got)
	}
}

// Twenty-Toed Toad: attacking with two creatures puts a +1/+1 counter on
// it and draws a card; attacking with twenty cards in hand wins.
func TestTwentyToedToad(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	toad := mhPush(g, me.ID, "Twenty-Toed Toad", "Creature — Frog Wizard", mhToadOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := mhMax(g, me); got != 20 {
		t.Errorf("maximum %d, want 20", got)
	}
	library := me.Library.Size()
	declareAttack(t, g, opp.ID, toad, bear)
	mhResolveAll(t, g)
	if n := findBattlefieldCardForTest(g, toad).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("+1/+1 counters on the Toad: %d, want 1", n)
	}
	if drawn := library - me.Library.Size(); drawn != 1 {
		t.Errorf("drew %d, want 1", drawn)
	}
	if g.State != game.StateActive {
		t.Fatal("the game ended with one counter and a small hand")
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	toad = mhPush(g, me.ID, "Twenty-Toed Toad", "Creature — Frog Wizard", mhToadOracle)
	for i := 0; me.Hand.Size() < 20; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	declareAttack(t, g, opp.ID, toad)
	mhResolveAll(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("outcome %+v, want a win for the Toad's controller", g.Outcome)
	}
}

// Anvil of Bogardan: no maximum for anyone, and each player's draw step
// draws an additional card and then asks that player to discard one.
func TestAnvilOfBogardan(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhPush(g, me.ID, "Anvil of Bogardan", "Artifact", mhAnvilOracle)
	for _, p := range g.Seats {
		if got := mhMax(g, p); got != game.NoMaxHandSize {
			t.Errorf("%s: %d, want no maximum", p.Name, got)
		}
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	library := opp.Library.Size()
	advanceToStepOf(t, g, 1, game.StepDraw)
	passPriorityAroundTable(t, g)
	if drawn := library - opp.Library.Size(); drawn != 2 {
		t.Errorf("the opponent drew %d in their draw step, want 2", drawn)
	}
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Errorf("the opponent owes %d discards, want 1", n)
	}
}

// Folio of Fancies: no maximum for anyone; {X}{X},{T} has each player
// draw X; {2}{U},{T} mills each opponent by their hand size.
func TestFolioOfFancies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	folio := mhPush(g, me.ID, "Folio of Fancies", "Artifact — Book", mhFolioOracle)
	if got := mhMax(g, g.Seats[1]); got != game.NoMaxHandSize {
		t.Errorf("an opponent's maximum %d, want no maximum", got)
	}
	libraries := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		libraries[p.ID] = p.Library.Size()
	}
	floatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, folio, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate {X}{X}: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if drawn := libraries[p.ID] - p.Library.Size(); drawn != 2 {
			t.Errorf("%s drew %d, want 2", p.Name, drawn)
		}
	}

	untapForTest(g, folio)
	opp := g.Seats[1]
	hand, library, mine := opp.Hand.Size(), opp.Library.Size(), me.Library.Size()
	floatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, folio, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the mill: %v", err)
	}
	passPriorityAroundTable(t, g)
	if milled := library - opp.Library.Size(); milled != hand {
		t.Errorf("an opponent with %d cards milled %d", hand, milled)
	}
	if me.Library.Size() != mine {
		t.Error("the Folio's controller milled")
	}
}

// Midnight Oil: enters with seven hour counters, so the maximum is
// seven; the draw step draws an extra card and removes two, so it is
// five; a discard costs a life.
func TestMidnightOil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oil := castCatalogSpell(t, g, "Midnight Oil", "Enchantment", mhMidnightOilOracle, nil)
	passPriorityAroundTable(t, g)
	if n := findBattlefieldCardForTest(g, oil).Counters["hour"]; n != 7 {
		t.Fatalf("hour counters %d, want 7", n)
	}
	if got := mhMax(g, me); got != 7 {
		t.Errorf("maximum %d, want 7", got)
	}
	seat := g.Turn.ActiveSeat
	advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	advanceToStepOf(t, g, seat, game.StepUpkeep)
	library := me.Library.Size()
	advanceToStepOf(t, g, seat, game.StepDraw)
	passPriorityAroundTable(t, g)
	if drawn := library - me.Library.Size(); drawn != 2 {
		t.Errorf("drew %d in the draw step, want 2", drawn)
	}
	if n := findBattlefieldCardForTest(g, oil).Counters["hour"]; n != 5 {
		t.Errorf("hour counters %d, want 5", n)
	}
	if got := mhMax(g, me); got != 5 {
		t.Errorf("maximum %d, want 5", got)
	}

	life := me.Life
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: me.ID, N: 1})
	})
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d after a discard, want one lost", life, me.Life)
	}
}

// Winter: no change without delirium; with four card types in your
// graveyard each opponent's maximum is three; a Reliquary Tower that
// entered after Winter wins (the 2024-09-20 ruling). The upkeep draws
// two for each player.
func TestWinterMisanthropicGuide(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mhPush(g, me.ID, "Winter, Misanthropic Guide", "Legendary Creature — Human Warlock", mhWinterOracle)
	if got := mhMax(g, opp); got != 7 {
		t.Errorf("without delirium: %d, want 7", got)
	}
	for _, typ := range []string{"Creature — Bear", "Instant", "Sorcery", "Artifact"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: typ, TypeLine: typ, Owner: me.ID, Controller: me.ID})
	}
	if got := mhMax(g, opp); got != 3 {
		t.Errorf("with four card types: %d, want 3", got)
	}
	if got := mhMax(g, me); got != 7 {
		t.Errorf("Winter's controller: %d, want 7", got)
	}
	mhTower(g, opp.ID)
	if got := mhMax(g, opp); got != game.NoMaxHandSize {
		t.Errorf("Winter, then the opponent's Reliquary Tower: %d, want no maximum", got)
	}

	libraries := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		libraries[p.ID] = p.Library.Size()
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	advanceToStepOf(t, g, 0, game.StepUpkeep)
	for _, p := range g.Seats {
		libraries[p.ID] = p.Library.Size()
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if drawn := libraries[p.ID] - p.Library.Size(); drawn != 2 {
			t.Errorf("%s drew %d at Winter's upkeep, want 2", p.Name, drawn)
		}
	}
}

// Register refuses an entry whose number is read at the time but whose
// kind is unknown, and accepts one with no N.
func TestDynamicHandSizeNeedsNoN(t *testing.T) {
	checkHandSize(Spec{Name: "ok", HandSize: []game.HandSizeStatic{{
		Kind:    game.HandSizeSet,
		Dynamic: func(*game.Game, *game.Card) (int, bool) { return 0, true },
	}}})
	defer func() {
		if recover() == nil {
			t.Error("an unknown kind with Dynamic was accepted")
		}
	}()
	checkHandSize(Spec{Name: "bad", HandSize: []game.HandSizeStatic{{
		Kind:    game.HandSizeKind(99),
		Dynamic: func(*game.Game, *game.Card) (int, bool) { return 0, true },
	}}})
}
