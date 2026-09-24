package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// game_end_cards_test.go — ADR 0057's first-wave cards (#749), each
// proved through the engine: a state-based check that does or does
// not take a player out, a game that does or does not end, and the
// outcome it records. One test per card at least; Herald of Eternal
// Dawn, the Aang deck's (#1306) card, gets the full set.

const (
	heraldOracle      = "4080f7e3-06d3-4d3d-9929-4e826cb66713"
	platinumAngelOrc  = "b148578c-c0bf-4785-b97c-4b6f83028008"
	persecutorOracle  = "7282f643-191b-41be-9f6f-82360c915d6d"
	felidarOracle     = "5ae3cbb9-9f0c-4077-ae7c-fba660d7fb4b"
	labManiacOracle   = "aa286dd5-aa19-446d-9003-684d81eb57ca"
	thassasOracleOrc  = "1de1b591-a73f-4974-b507-8c63e07a0868"
	angelsGraceOracle = "66ca8a60-e028-4a5f-8177-860b888cb9d1"
)

// newTwoSeatCatalogGame is newCatalogGame at a two-seat table, where a
// single departure ends the game.
func newTwoSeatCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// checkState runs the state-based check the way play does: the table
// walks one step, and the step boundary checks.
func checkState(t *testing.T, g *game.Game) {
	t.Helper()
	if g.State != game.StateActive {
		return
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}

func eventsOfKind(g *game.Game, kind game.EventKind) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

func destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

func emptyLibrary(p *game.Player) {
	for p.Library.Size() > 0 {
		_, _ = p.Library.PopTop()
	}
}

// --- Herald of Eternal Dawn ----------------------------------------

// TestHeraldOfEternalDawnStopsEveryLossForItsController is the static's
// "you can't lose" half, for every CR 104.3 route in one table: life,
// poison and commander damage on its controller stay in; the same
// numbers on an opponent do not.
func TestHeraldOfEternalDawnStopsEveryLossForItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, me.ID, "Herald of Eternal Dawn", heraldOracle, "Creature — Angel")
	me.Life = -3
	me.Poison = game.PoisonLethal
	me.CommanderDamage = map[uuid.UUID]int{uuid.New(): 25}
	opp.Life = 0

	checkState(t, g)
	checkState(t, g)
	if me.Eliminated {
		t.Fatalf("the Herald's controller lost")
	}
	if !opp.Eliminated {
		t.Fatalf("an opponent at 0 life stayed in — the Herald protects only its controller")
	}
	var causes []game.LossCause
	g.ReadSnapshot(func() { causes = g.CantLoseCausesForEffect(me) })
	if len(causes) != len(game.GatableLossCauses) {
		t.Errorf("cant_lose %v", causes)
	}

	// An effect loss is stopped at the moment it happens.
	g.WithWriteLock(func() {
		if lost, _ := g.LoseTheGameForEffect(me.ID, uuid.Nil); lost {
			t.Fatalf("lost to an effect under the Herald")
		}
	})
}

// TestHeraldOfEternalDawnLeavingLetsTheLossLand: the gate is read at
// every check, so the first check after the Herald dies takes its
// controller out — to life, the first cause in CR 704.5 order.
func TestHeraldOfEternalDawnLeavingLetsTheLossLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	herald := pushPermanentForTest(g, me.ID, "Herald of Eternal Dawn", heraldOracle, "Creature — Angel")
	me.Life = -3
	checkState(t, g)
	if me.Eliminated {
		t.Fatalf("lost with the Herald out")
	}
	destroy(t, g, herald)
	checkState(t, g)
	if !me.Eliminated {
		t.Fatalf("still in at -3 life after the Herald died")
	}
	elim := eventsOfKind(g, game.EventPlayerEliminated)
	if len(elim) != 1 || elim[0].Label != string(game.LossLife) {
		t.Errorf("eliminations %+v, want one for life", elim)
	}
}

// TestHeraldOfEternalDawnForgetsAnOldEmptyDraw: a draw from an empty
// library under the Herald is not held against its controller once the
// Herald is gone (CR 704.5b: only draws since the last check).
func TestHeraldOfEternalDawnForgetsAnOldEmptyDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	herald := pushPermanentForTest(g, me.ID, "Herald of Eternal Dawn", heraldOracle, "Creature — Angel")
	emptyLibrary(me)
	if err := g.DrawCard(me.ID); err == nil {
		t.Logf("draw from an empty library returned no error")
	}
	checkState(t, g)
	if me.Eliminated || me.AttemptedEmptyDraw {
		t.Fatalf("eliminated %v, flag %v after the stopped check", me.Eliminated, me.AttemptedEmptyDraw)
	}
	destroy(t, g, herald)
	checkState(t, g)
	if me.Eliminated {
		t.Fatalf("lost to a draw made under the Herald")
	}
}

// TestHeraldOfEternalDawnStopsAnOpponentsWin: "your opponents can't
// win the game" — an opponent's Felidar Sovereign triggers, and the
// win is prevented and logged, and the game goes on.
func TestHeraldOfEternalDawnStopsAnOpponentsWin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herald := pushPermanentForTest(g, me.ID, "Herald of Eternal Dawn", heraldOracle, "Creature — Angel")
	felidar := pushPermanentForTest(g, opp.ID, "Felidar Sovereign", felidarOracle, "Creature — Cat Beast")
	opp.Life = 45
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatalf("the game ended past the Herald: %+v", g.Outcome)
	}
	prevented := eventsOfKind(g, game.EventWinPrevented)
	if len(prevented) != 1 || prevented[0].Actor != opp.ID || prevented[0].Source != felidar || prevented[0].Target != herald {
		t.Fatalf("win_prevented %+v", prevented)
	}
}

// TestHeraldOfEternalDawnHasFlashAndFlying: the printed keywords.
func TestHeraldOfEternalDawnHasFlashAndFlying(t *testing.T) {
	spec, ok := Lookup(heraldOracle)
	if !ok {
		t.Fatalf("Herald of Eternal Dawn is not in the catalog")
	}
	want := map[string]bool{"flash": true, "flying": true}
	for _, k := range spec.PrintedKeywords {
		delete(want, k)
	}
	if len(want) != 0 {
		t.Errorf("missing keywords %v", want)
	}
}

// --- Platinum Angel ---------------------------------------------------

// TestPlatinumAngelConcedeStillLoses: a two-seat game, the Angel's
// controller at -5 stays in; when they concede, they lose (CR 104.3a)
// and the opponent wins as the last player standing.
func TestPlatinumAngelConcedeStillLoses(t *testing.T) {
	g := newTwoSeatCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Platinum Angel", platinumAngelOrc, "Artifact Creature — Angel")
	me.Life = -5
	checkState(t, g)
	if me.Eliminated {
		t.Fatalf("lost under the Angel")
	}
	if err := g.Concede(me.ID); err != nil {
		t.Fatal(err)
	}
	if g.Outcome == nil || g.Outcome.Winner != opp.ID || g.Outcome.Cause != game.OutcomeCauseLastStanding {
		t.Fatalf("outcome %+v", g.Outcome)
	}
}

// --- Abyssal Persecutor -------------------------------------------------

// TestAbyssalPersecutorKeepsOpponentsInAndCantWin: its opponent at -5
// stays in; its controller's Felidar Sovereign can't win; when the
// Persecutor dies, the opponent loses at the next check with no window,
// and the controller wins as the last player standing.
func TestAbyssalPersecutorKeepsOpponentsInAndCantWin(t *testing.T) {
	g := newTwoSeatCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	persecutor := pushPermanentForTest(g, me.ID, "Abyssal Persecutor", persecutorOracle, "Creature — Demon")
	opp.Life = -5
	checkState(t, g)
	if opp.Eliminated {
		t.Fatalf("the Persecutor's opponent lost at -5")
	}
	g.WithWriteLock(func() {
		if won, _ := g.WinTheGameForEffect(me.ID, uuid.Nil); won {
			t.Fatalf("the Persecutor's controller won by effect")
		}
	})
	destroy(t, g, persecutor)
	checkState(t, g)
	if !opp.Eliminated {
		t.Fatalf("the opponent survived the Persecutor leaving")
	}
	if g.Outcome == nil || g.Outcome.Winner != me.ID || g.Outcome.Cause != game.OutcomeCauseLastStanding {
		t.Fatalf("outcome %+v", g.Outcome)
	}
}

// --- Felidar Sovereign -----------------------------------------------

// TestFelidarSovereignWinsAFourPlayerGameAtItsUpkeep is test plan item
// 2 with the real card: 40 life at the upkeep wins, and the other three
// seats are still seated when the game ends.
func TestFelidarSovereignWinsAFourPlayerGameAtItsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	felidar := pushPermanentForTest(g, me.ID, "Felidar Sovereign", felidarOracle, "Creature — Cat Beast")
	me.Life = 40
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	want := game.GameOutcome{Kind: game.OutcomeWin, Winner: me.ID, Cause: game.OutcomeCauseEffect, Source: felidar}
	if g.State != game.StateEnded || g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("state %s outcome %+v, want %+v", g.State, g.Outcome, want)
	}
	for _, p := range g.Seats {
		if p.Eliminated {
			t.Errorf("%s was eliminated by an effect win", p.Name)
		}
	}
	if n := len(eventsOfKind(g, game.EventEffectError)); n != 0 {
		t.Errorf("%d effect errors", n)
	}
}

// TestFelidarSovereignWinsATwoPlayerGame: the same at two seats.
func TestFelidarSovereignWinsATwoPlayerGame(t *testing.T) {
	g := newTwoSeatCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Felidar Sovereign", felidarOracle, "Creature — Cat Beast")
	me.Life = 52
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("outcome %+v", g.Outcome)
	}
}

// TestFelidarSovereignInterveningIf: CR 603.4, both halves. Below 40
// the trigger never fires; at 40 it fires, and if life drops below 40
// before it resolves, it does nothing.
func TestFelidarSovereignInterveningIf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	felidar := pushPermanentForTest(g, me.ID, "Felidar Sovereign", felidarOracle, "Creature — Cat Beast")
	me.Life = 39
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, felidar) != nil {
		t.Fatalf("triggered at 39 life")
	}
	passPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatalf("won at 39 life")
	}

	// Walk off this upkeep to the same seat's next one.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	me.Life = 40
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, felidar) == nil {
		t.Fatalf("did not trigger at 40 life")
	}
	me.Life = 37 // a Bolt in response
	passPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatalf("won after dropping to 37 in response — the intervening if is checked again on resolution")
	}
}

// --- Laboratory Maniac --------------------------------------------------

// TestLaboratoryManiacWinsInsteadOfDrawing: a draw from an empty
// library wins the game.
func TestLaboratoryManiacWinsInsteadOfDrawing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	lab := pushPermanentForTest(g, me.ID, "Laboratory Maniac", labManiacOracle, "Creature — Human Wizard")
	emptyLibrary(me)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	want := game.GameOutcome{Kind: game.OutcomeWin, Winner: me.ID, Cause: game.OutcomeCauseEffect, Source: lab}
	if g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("outcome %+v, want %+v", g.Outcome, want)
	}
	if me.AttemptedEmptyDraw {
		t.Errorf("the replaced draw set the empty-draw flag")
	}
}

// TestLaboratoryManiacAgainstPlatinumAngel is test plan item 10: an
// opponent's Angel prevents the win, the draw is still replaced, and
// the Maniac's controller does not lose for it at the next check (the
// 2021-03-19 ruling).
func TestLaboratoryManiacAgainstPlatinumAngel(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, me.ID, "Laboratory Maniac", labManiacOracle, "Creature — Human Wizard")
	angel := pushPermanentForTest(g, opp.ID, "Platinum Angel", platinumAngelOrc, "Artifact Creature — Angel")
	emptyLibrary(me)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if g.State != game.StateActive {
		t.Fatalf("the game ended past the Angel: %+v", g.Outcome)
	}
	prevented := eventsOfKind(g, game.EventWinPrevented)
	if len(prevented) != 1 || prevented[0].Target != angel {
		t.Fatalf("win_prevented %+v", prevented)
	}
	if me.AttemptedEmptyDraw {
		t.Fatalf("the prevented win left the draw unreplaced")
	}
	checkState(t, g)
	if me.Eliminated {
		t.Fatalf("lost for a draw the Maniac replaced")
	}
}

// --- Thassa's Oracle --------------------------------------------------

// castThassasOracle puts a blue permanent with {U}{U} out for devotion,
// casts the Oracle, and passes until its ETB asks the pick. Returns the
// Oracle's instance ID.
func castThassasOracle(t *testing.T, g *game.Game, devotion string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Blue Devotion", ManaCost: devotion,
		TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	id := castCatalogSpell(t, g, "Thassa's Oracle", "Creature — Merfolk Wizard", thassasOracleOrc, nil)
	passPriorityAroundTable(t, g)
	return id
}

// TestThassasOracleWinsWhenDevotionCoversTheLibrary: devotion 2, two
// cards left. The pick keeps one on top, the other goes under, and 2 >=
// 2 wins.
func TestThassasOracleWinsWhenDevotionCoversTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for me.Library.Size() > 2 {
		_, _ = me.Library.PopTop()
	}
	oracle := castThassasOracle(t, g, "{U}{U}")
	pick := latestRevealPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick prompt: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 2 {
		t.Fatalf("pick %d..%d over %d cards, want up to one of two", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, pick.ChooseCards[1:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	want := game.GameOutcome{Kind: game.OutcomeWin, Winner: me.ID, Cause: game.OutcomeCauseEffect, Source: oracle}
	if g.Outcome == nil || *g.Outcome != want {
		t.Fatalf("outcome %+v, want %+v", g.Outcome, want)
	}
	if n := len(eventsOfKind(g, game.EventEffectError)); n != 0 {
		t.Errorf("%d effect errors", n)
	}
}

// TestThassasOracleArrangesAndDoesNotWinShortOfTheLibrary: devotion 2
// against a bigger library — the kept card is on top, the other is on
// the bottom, and nobody wins.
func TestThassasOracleArrangesAndDoesNotWinShortOfTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castThassasOracle(t, g, "{U}")
	// Devotion: the Oracle has no printed cost in this harness, so X is
	// the one {U} seeded above.
	pick := latestRevealPickFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 1 {
		t.Fatalf("pick %+v", pick)
	}
	kept := pick.ChooseCards[0]
	if err := g.ResolveRevealPick(pick.ID, me.ID, []uuid.UUID{}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if g.State != game.StateActive {
		t.Fatalf("won with X below the library size")
	}
	bottom := me.Library.Cards[0].InstanceID
	if bottom != kept {
		t.Errorf("the card not kept is not on the bottom")
	}
}

// --- Angel's Grace ----------------------------------------------------

// TestAngelsGraceLastsTheTurn is test plan item 14 with the real card:
// this turn the caster can't lose and an opponent can't win; at the
// next turn the grant is gone and a caster left at 0 life loses.
func TestAngelsGraceLastsTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Angel's Grace", "Instant", angelsGraceOracle, nil)
	passPriorityAroundTable(t, g)
	me.Life = 0
	checkState(t, g)
	if me.Eliminated {
		t.Fatalf("lost this turn under Angel's Grace")
	}
	g.WithWriteLock(func() {
		if won, _ := g.WinTheGameForEffect(opp.ID, uuid.Nil); won {
			t.Fatalf("an opponent won this turn")
		}
	})
	var gates []game.GameEndGateSource
	g.ReadSnapshot(func() { gates = g.GameEndGatesForEffect(me) })
	if len(gates) != 1 || !gates[0].ThisTurn || gates[0].SourceName != "Angel's Grace" {
		t.Errorf("end gates %+v", gates)
	}
	advanceToUpkeepOf(t, g, 1)
	if !me.Eliminated {
		t.Fatalf("still in at 0 life after the turn Angel's Grace was cast")
	}
}

// --- Pact of Negation (the effect loss, now immediate) ----------------

// pactDebtDue casts a Pact of Negation for seat 0 and walks to its next
// upkeep, where the debt's pay-or-lose prompt is open.
func pactDebtDue(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[0]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	pact := handCardFull(me, "Pact of Negation", "Instant", "{0}", pactOfNegationOracle, []string{"U"})
	if err := g.CastSpell(me.ID, pact, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Pact of Negation: %v", err)
	}
	passPriorityAroundTable(t, g)
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
}

// TestPactOfNegationLossIsImmediateAndMovesTheTurnOn: ADR 0057
// Decision 3. Declining in a four-seat game takes the active player out
// at once, with an "effect" cause and no effect error, and the next
// state check moves the turn to the next seat.
func TestPactOfNegationLossIsImmediateAndMovesTheTurnOn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pactDebtDue(t, g)
	answerPayUnless(t, g, me.ID, false)
	if !me.Eliminated {
		t.Fatalf("an unpaid Pact did not lose its controller the game")
	}
	elim := eventsOfKind(g, game.EventPlayerEliminated)
	if len(elim) != 1 || elim[0].Label != string(game.LossEffect) {
		t.Fatalf("eliminations %+v, want one effect loss", elim)
	}
	if n := len(eventsOfKind(g, game.EventEffectError)); n != 0 {
		t.Errorf("%d effect errors; the stop is clean", n)
	}
	if g.State != game.StateActive {
		t.Fatalf("a four-seat game ended on one loss")
	}
	if g.Turn.ActiveSeat == 0 {
		t.Errorf("the turn did not move on from the departed active player")
	}
}

// TestPactOfNegationUnderPlatinumAngel: the gate is read when the loss
// happens, so an Angel on the battlefield then stops it for good.
func TestPactOfNegationUnderPlatinumAngel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	angel := pushPermanentForTest(g, me.ID, "Platinum Angel", platinumAngelOrc, "Artifact Creature — Angel")
	pactDebtDue(t, g)
	answerPayUnless(t, g, me.ID, false)
	if me.Eliminated {
		t.Fatalf("lost to the Pact under Platinum Angel")
	}
	destroy(t, g, angel)
	checkState(t, g)
	if me.Eliminated {
		t.Fatalf("the stopped loss was remembered (CR 614.17a: a can't doesn't reach back)")
	}
}
