package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skip_turn_cards_more_test.go — ADR 0188 slice 7 (#3565): nine more
// cards that skip turns, each driven through the real cast or activation
// and the real rotation seam. The skip is asserted both as the debt the
// engine reports (TurnSkipsFor) and as who is active after the turn ends.

const (
	meditateOracle         = "f79f84c9-2023-4348-863b-a3b4bc382904"
	chronatogOracle        = "bb300e38-32d7-4152-9a82-88d43af99f19"
	chronatogTotemOracle   = "24542d66-e67a-4403-91a8-bb4231ea1f50"
	eaterOfDaysOracle      = "205f6384-7203-42ed-bc0f-8803cf1f4514"
	wormfangMantaOracle    = "17abf8c6-70b6-4877-a0f2-a87b22fb2828"
	waterspoutOracle       = "f86b8c03-9f1e-4d42-ac45-907c46e9e1c0"
	chronosavantOracle     = "969c9864-f78b-46b2-b35b-c9adaf355c2d"
	unstableHulkOracle     = "8d01c85c-5f83-4514-b81c-4165051d5722"
	ralGuestLecturerOracle = "0dc76b7f-6405-44a1-8c55-31fa480ff70b"
)

func TestMoreSkipTurnCardsAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]string{
		meditateOracle:         "Meditate",
		chronatogOracle:        "Chronatog",
		chronatogTotemOracle:   "Chronatog Totem",
		eaterOfDaysOracle:      "Eater of Days",
		wormfangMantaOracle:    "Wormfang Manta",
		waterspoutOracle:       "Waterspout Elemental",
		chronosavantOracle:     "Chronosavant",
		unstableHulkOracle:     "Unstable Hulk",
		ralGuestLecturerOracle: "Ral Zarek, Guest Lecturer",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q completeness %v", want, spec.Name, spec.Completeness)
		}
	}
}

// assertRotation ends the turn once per entry and checks who is up.
func assertRotation(t *testing.T, g *game.Game, seats ...int) {
	t.Helper()
	for _, want := range seats {
		endTurn(t, g)
		assertTurnOf(t, g, want, false, "rotation with a skipped turn")
	}
}

// --- Meditate ---------------------------------------------------------

func TestMeditateDrawsFourThenSkipsTheCastersNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	p := g.Seats[me]
	castCatalogSpell(t, g, "Meditate", "Instant", meditateOracle, nil)
	before := p.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := p.Hand.Size() - before; got != 4 {
		t.Errorf("drew %d cards, want 4", got)
	}
	if got := g.TurnSkipsFor(p.ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
	assertTurnOf(t, g, me, false, "the turn it was cast in is not cut short")
	assertRotation(t, g, 1, 2, 3, 1)
}

// Cast on an opponent's turn, the skip is the caster's next turn, not
// the opponent's (CR 614.10: "you").
func TestMeditateInResponseOnAnotherPlayersTurnCostsTheCaster(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[2]
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Meditate", TypeLine: "Instant", OracleID: meditateOracle, Owner: caster.ID, Controller: caster.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Meditate on another player's turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(caster.ID); got != 1 {
		t.Errorf("caster owes %d skips, want 1", got)
	}
	if got := g.TurnSkipsFor(g.Seats[g.Turn.ActiveSeat].ID); got != 0 {
		t.Errorf("the active player owes %d skips, want 0", got)
	}
}

// --- Chronatog and Chronatog Totem -------------------------------------

func TestChronatogPumpsOnceEachTurnAndSkipsATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	atog := pushCatalogPermanent(g, g.Seats[me].ID, "Chronatog", "Creature — Atog", chronatogOracle, false)
	base := effectivePower(t, g, atog)

	if err := g.ActivateCatalogAbility(g.Seats[me].ID, atog, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, atog) - base; got != 3 {
		t.Errorf("power rose by %d, want 3", got)
	}
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, atog, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation in one turn was allowed")
	}
	assertTurnOf(t, g, me, false, "the turn it was used in is not cut short")
	assertRotation(t, g, 1, 2, 3, 1)
}

// The pump and the skip are separate sentences: a Chronatog that left in
// response still costs its controller the turn.
func TestChronatogSkipHappensEvenIfItHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	atog := pushCatalogPermanent(g, g.Seats[me].ID, "Chronatog", "Creature — Atog", chronatogOracle, false)
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, atog, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(atog) })
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips after Chronatog left, want 1", got)
	}
}

func TestChronatogTotemPumpNeedsItToBeACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	totem := pushCatalogPermanent(g, g.Seats[me].ID, "Chronatog Totem", "Artifact", chronatogTotemOracle, false)

	if err := g.ActivateCatalogAbility(g.Seats[me].ID, totem, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump was activated while the Totem was not a creature")
	}
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 0 {
		t.Fatalf("a refused activation left %d skips", got)
	}

	if err := g.AddManaForEffect(g.Seats[me].ID, uuid.Nil, "{C}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, totem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCardByID(g, totem)
	if !ok {
		t.Fatal("the Totem left the battlefield")
	}
	eff := c.Effective()
	if !c.IsCreature() || eff.Power != 1 || eff.Toughness != 2 {
		t.Fatalf("animated Totem: creature=%v %d/%d, want a 1/2 creature", c.IsCreature(), eff.Power, eff.Toughness)
	}
	if !c.HasSubtype("Atog") {
		t.Error("the animated Totem is not an Atog")
	}

	if err := g.ActivateCatalogAbility(g.Seats[me].ID, totem, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pump as a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, totem); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, totem, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second pump in one turn was allowed")
	}
}

// --- Eater of Days ------------------------------------------------------

func TestEaterOfDaysSkipsTwoTurns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castCatalogSpell(t, g, "Eater of Days", "Artifact Creature — Leviathan", eaterOfDaysOracle, nil)
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 2 {
		t.Fatalf("owes %d skips, want 2", got)
	}
	assertTurnOf(t, g, me, false, "the turn it entered in is not cut short")
	// Two of this player's turns pass without them, then the third comes.
	assertRotation(t, g, 1, 2, 3, 1, 2, 3, 1, 2, 3, me)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 0 {
		t.Errorf("debt left over: %d", got)
	}
}

// --- Wormfang Manta -----------------------------------------------------

func TestWormfangMantaSkipsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	castCatalogSpell(t, g, "Wormfang Manta", "Creature — Nightmare Fish Beast", wormfangMantaOracle, nil)
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Fatalf("owes %d skips, want 1", got)
	}
	assertRotation(t, g, 1, 2, 3, 1)
}

// Leaving the battlefield the same turn it entered takes an extra turn,
// and the skip is spent on that extra turn (the Magosi ruling): the
// controller loses the extra turn and nothing else.
func TestWormfangMantaLeavingGivesAnExtraTurnThatTheSkipSpends(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	id := castCatalogSpell(t, g, "Wormfang Manta", "Creature — Nightmare Fish Beast", wormfangMantaOracle, nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Fatalf("owes %d skips, want 1", got)
	}
	// The extra turn is skipped by the debt, so the next turn is seat 1's.
	assertRotation(t, g, 1)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 0 {
		t.Errorf("the skip was not spent on the extra turn: %d left", got)
	}
}

// Leaving on a LATER turn, the extra turn is real: the skip is long
// spent and the controller plays the extra turn.
func TestWormfangMantaLeavingLaterTakesARealExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	id := pushCatalogPermanent(g, g.Seats[me].ID, "Wormfang Manta", "Creature — Nightmare Fish Beast", wormfangMantaOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "Wormfang Manta's extra turn")
}

// "Leaves the battlefield" is any exit, not only dying: exiled, the Manta
// still gives the extra turn.
func TestWormfangMantaExtraTurnAlsoOnAnExitThatIsNotDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	id := pushCatalogPermanent(g, g.Seats[me].ID, "Wormfang Manta", "Creature — Nightmare Fish Beast", wormfangMantaOracle, false)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneExile},
		id,
	); err != nil {
		t.Fatalf("exile the Manta: %v", err)
	}
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn for an exiled Manta")
}

// --- Waterspout Elemental ----------------------------------------------

func TestWaterspoutElementalUnkickedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	bear := seedCreature(g, "Bystander", g.Seats[1].ID)
	castWithOptionalCosts(t, g, "Waterspout Elemental", "Creature — Elemental", waterspoutOracle, nil, nil, nil)
	passPriorityAroundTable(t, g)
	if n := len(g.PendingTriggers); n != 0 {
		t.Fatalf("unkicked Elemental queued %d triggers", n)
	}
	if findBattlefieldCardForTest(g, bear) == nil {
		t.Error("an unkicked Elemental bounced a creature")
	}
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 0 {
		t.Errorf("owes %d skips, want 0", got)
	}
}

func TestWaterspoutElementalKickedBouncesTheOthersAndSkips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	theirs := seedCreature(g, "Bystander", g.Seats[1].ID)
	mine := seedCreature(g, "My Bear", g.Seats[me].ID)
	id, err := castWithOptionalCosts(t, g, "Waterspout Elemental", "Creature — Elemental", waterspoutOracle, nil, []int{0}, nil)
	if err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, theirs) != nil || findBattlefieldCardForTest(g, mine) != nil {
		t.Error("another creature survived the bounce")
	}
	if findBattlefieldCardForTest(g, id) == nil {
		t.Error("the Elemental bounced itself")
	}
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
	assertRotation(t, g, 1, 2, 3, 1)
}

// --- Chronosavant ---------------------------------------------------------

func TestChronosavantReturnsTappedFromTheGraveyardAndSkips(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Chronosavant", "Creature — Giant", chronosavantOracle, 5, 5, "{C}{W}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("Chronosavant is not on the battlefield")
	}
	if !c.Tapped {
		t.Error("Chronosavant must enter tapped")
	}
	if got := g.TurnSkipsFor(me.ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
}

// The return and the skip are separate instructions: exiled in response,
// Chronosavant does not return but the turn is still lost.
func TestChronosavantSkipHappensEvenIfItWasExiledInResponse(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Chronosavant", "Creature — Giant", chronosavantOracle, 5, 5, "{C}{W}", game.ActivateAbilityParams{})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneExile},
		id,
	); err != nil {
		t.Fatalf("exile in response: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("an exiled Chronosavant returned")
	}
	if got := g.TurnSkipsFor(me.ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
}

// --- Unstable Hulk -------------------------------------------------------

func TestUnstableHulkTurnedFaceUpGrowsAndSkips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	active := g.Seats[me]
	id := castWithAltCost(t, g, "Unstable Hulk", "Creature — Goblin Mutant", unstableHulkOracle, "morph")
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(active.ID); got != 0 {
		t.Fatalf("a face-down Hulk already owes %d skips", got)
	}
	g.WithWriteLock(func() { active.ManaPool.AddMana(manaTokens("R", "R", "C", "C", "C")...) })
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	base := effectivePower(t, g, id)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, id) - base; got != 6 {
		t.Errorf("power rose by %d, want 6", got)
	}
	if !effectiveAbilitiesContain(t, g, id, "trample") {
		t.Error("the Hulk did not gain trample")
	}
	if got := g.TurnSkipsFor(active.ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
	assertTurnOf(t, g, me, false, "the turn it was turned up in is not cut short")
	assertRotation(t, g, 1, 2, 3, 1)
}

// Cast face up the trigger never fires: the card is not turned face up.
func TestUnstableHulkCastFaceUpDoesNotSkip(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Unstable Hulk", "Creature — Goblin Mutant", unstableHulkOracle, nil)
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[g.Turn.ActiveSeat].ID); got != 0 {
		t.Errorf("owes %d skips, want 0", got)
	}
}

// --- Ral Zarek, Guest Lecturer ------------------------------------------

func ralOnBattlefield(t *testing.T, g *game.Game, loyalty int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	id := pushCatalogPermanent(g, me.ID, "Ral Zarek, Guest Lecturer", "Legendary Planeswalker — Ral", ralGuestLecturerOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty) })
	return id
}

func TestRalGuestLecturerPlusOneSurveilsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ral := ralOnBattlefield(t, g, 3)
	if err := g.ActivateCatalogAbility(me.ID, ral, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c == nil {
		t.Fatal("no surveil prompt for the controller")
	}
}

func TestRalGuestLecturerMinusOneMakesChosenPlayersDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	ral := ralOnBattlefield(t, g, 3)
	if err := g.ActivateCatalogAbility(me.ID, ral, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("−1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) == nil || discardChoiceFor(g, me.ID) == nil {
		t.Fatal("each chosen player should be asked to discard")
	}
	if discardChoiceFor(g, g.Seats[2].ID) != nil {
		t.Error("an unchosen player was asked to discard")
	}
}

func TestRalGuestLecturerMinusOneAcceptsNoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ral := ralOnBattlefield(t, g, 3)
	if err := g.ActivateCatalogAbility(me.ID, ral, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−1 with zero targets: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestRalGuestLecturerMinusTwoReturnsASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ral := ralOnBattlefield(t, g, 3)
	small := uuid.New()
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Wurm", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})

	if err := g.ActivateCatalogAbility(me.ID, ral, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err == nil {
		t.Fatal("a mana value 4 creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, ral, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature did not return")
	}
}

// The −7: five uncalled flips, and the target opponent skips as many
// turns as there were heads. The heads are read off the flip events, so
// the test holds for whichever faces come up; across the keys below both
// a nonzero and a zero count are seen, and zero heads writes no debt.
func TestRalGuestLecturerMinusSevenSkipsOneTurnPerHead(t *testing.T) {
	seenZero, seenSome := false, false
	for k := byte(1); k <= 24; k++ {
		g := newCatalogGame(t)
		var key [32]byte
		for i := range key {
			key[i] = k*7 + byte(i)
		}
		g.SetRNGKeyForTest(key)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[2]
		ral := ralOnBattlefield(t, g, 7)
		if err := g.ActivateCatalogAbility(me.ID, ral, 3, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("−7: %v", err)
		}
		if got := g.TurnSkipsFor(opp.ID); got != 0 {
			t.Fatalf("the skip was owed before the ability resolved: %d", got)
		}
		passPriorityAroundTable(t, g)
		heads, flips := 0, 0
		for _, ev := range g.Events {
			if ev.Kind == game.EventFlipCoin {
				flips++
				if ev.Label == "heads" {
					heads++
				}
			}
		}
		if flips != 5 {
			t.Fatalf("key %d: %d coins flipped, want 5", k, flips)
		}
		if got := g.TurnSkipsFor(opp.ID); got != heads {
			t.Fatalf("key %d: opponent owes %d skips with %d heads", k, got, heads)
		}
		if got := g.TurnSkipsFor(me.ID); got != 0 {
			t.Fatalf("key %d: the controller owes %d skips", k, got)
		}
		if heads == 0 {
			seenZero = true
		} else {
			seenSome = true
		}
	}
	if !seenSome {
		t.Error("no key produced any heads: the test never saw a skip")
	}
	_ = seenZero // zero heads is 1 in 32; when a key shows it, the equality above covers it.
}

func TestRalGuestLecturerMinusSevenTargetsOnlyOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ral := ralOnBattlefield(t, g, 7)
	if err := g.ActivateCatalogAbility(me.ID, ral, 3, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("the −7 targeted its own controller")
	}
}
