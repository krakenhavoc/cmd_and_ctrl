package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_for_cards_test.go — S22's passive draw engines that spend
// life instead of mana: Necropotence, Yawgmoth's Bargain,
// Griselbrand, Vilis and Bloodgift Demon.
//
// Three things these tests are actually about, none of which a
// "hand grew by N" assertion would catch:
//
//   - "Skip your draw step" must skip a STEP and only its
//     controller's. The observable difference is the turn cursor,
//     not the hand size.
//   - Necropotence's exile is FACE DOWN. The assertion that matters
//     is that no seat — including the controller — is a knower of
//     the card while it sits in exile, because exile is a public
//     zone and every other exile in the engine marks the whole
//     table.
//   - The delayed delivery is "at the beginning of YOUR next end
//     step". An opponent's end step in between must not hand the
//     card over early.

const (
	necropotenceOracle     = "94a844d2-0574-45a7-b347-e0e329767c42"
	yawgmothsBargainOracle = "f7f76f39-a0de-4bda-86b6-0f291892fcec"
	griselbrandOracle      = "f759d112-76db-4091-a22b-b9f19ab6fa5f"
	vilisOracle            = "7c209753-121b-4859-944b-4e33f885e777"
	bloodgiftDemonOracle   = "e64b184d-9746-4723-a928-d459b5c3ee6c"
)

// --- helpers ------------------------------------------------------

// pushLibraryTopForTest puts a named card on TOP of a player's
// library — the end ExileTopFaceDown and every draw read from.
// pushLibraryCardForTest is the bottom-pushing sibling used by the
// tutor tests, where "first match" wants determinism rather than a
// known top card.
func pushLibraryTopForTest(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// soleHandCardForTest empties a player's hand and leaves one named
// card in it, so DiscardRandomForEffect's random pick is the card
// the test wants to watch. The catalog has no "discard this specific
// card" mutator outside an additional cost, and the trigger under
// test does not care which card it was.
func soleHandCardForTest(p *game.Player, name string) uuid.UUID {
	p.Hand.Cards = nil
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Instant",
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// findInZone returns the card with the given instance ID from the
// zone, and whether it was there.
func findInZone(z *game.Zone, cardID uuid.UUID) (game.Card, bool) {
	if z == nil {
		return game.Card{}, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == cardID {
			return c, true
		}
	}
	return game.Card{}, false
}

// walkSteps advances the turn cursor up to `limit` observable steps,
// recording every (seat, step) pair it lands on, and stops early
// once `done` is satisfied. Returns the trace.
//
// A trace rather than a spot check because "the draw step was
// skipped" is a statement about a step that never happened: there is
// no moment at which a test can look at the cursor and see it.
func walkSteps(t *testing.T, g *game.Game, limit int, done func() bool) []stepFrame {
	t.Helper()
	trace := []stepFrame{{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step}}
	for i := 0; i < limit; i++ {
		if done() {
			return trace
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep iter %d: %v", i, err)
		}
		trace = append(trace, stepFrame{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step})
	}
	t.Fatalf("turn cursor never reached the wanted position in %d steps", limit)
	return nil
}

// stepFrame is one position of the turn cursor.
type stepFrame struct {
	Seat int
	Step game.Step
}

// sawStep reports whether the trace ever landed on a given seat's
// given step.
func sawStep(trace []stepFrame, seat int, step game.Step) bool {
	for _, f := range trace {
		if f.Seat == seat && f.Step == step {
			return true
		}
	}
	return false
}

// --- Skip your draw step ------------------------------------------

// TestNecropotenceSkipsOnlyItsControllersDrawStep — the replacement
// cancels the step-transition event for seat 1 and leaves every
// other seat's draw step alone.
//
// Seat 1 rather than seat 0 on purpose: CR 103.8a already skips the
// starting player's turn-1 draw step, so asserting on seat 0's first
// turn would pass with the card doing nothing at all.
func TestNecropotenceSkipsOnlyItsControllersDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	_ = seedReplacementPermanent(g, necropotenceOracle, "Necropotence", owner.ID)

	handBefore := owner.Hand.Size()
	trace := walkSteps(t, g, 200, func() bool {
		return g.Turn.ActiveSeat == 2 && g.Turn.Step == game.StepPrecombatMain
	})

	if sawStep(trace, 1, game.StepDraw) {
		t.Errorf("seat 1's draw step happened despite Necropotence")
	}
	if !sawStep(trace, 2, game.StepDraw) {
		t.Errorf("seat 2's draw step was skipped too — the replacement is not seat-scoped")
	}
	if owner.Hand.Size() != handBefore {
		t.Errorf("Necropotence's controller drew %d cards through the skipped step",
			owner.Hand.Size()-handBefore)
	}
	if seat2 := g.Seats[2]; seat2.Hand.Size() == 0 {
		t.Errorf("seat 2 has an empty hand; the walk never reached their turn")
	}
}

// TestYawgmothsBargainSkipsTheDrawStepToo — the same helper on the
// other card that prints the clause. Cheap, and it is the test that
// fails if SkipYourDrawStep is ever inlined into one card file and
// diverges.
func TestYawgmothsBargainSkipsTheDrawStepToo(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	_ = seedReplacementPermanent(g, yawgmothsBargainOracle, "Yawgmoth's Bargain", owner.ID)

	trace := walkSteps(t, g, 200, func() bool {
		return g.Turn.ActiveSeat == 2 && g.Turn.Step == game.StepPrecombatMain
	})
	if sawStep(trace, 1, game.StepDraw) {
		t.Errorf("seat 1's draw step happened despite Yawgmoth's Bargain")
	}
}

// TestSkipDrawStepEndsWithThePermanent — the replacement is a
// continuous effect from a permanent (CR 113.6), so removing the
// permanent restores the draw step. Guards against anyone
// "simplifying" it into a flag on the player.
func TestSkipDrawStepEndsWithThePermanent(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", owner.ID)

	if _, err := g.Battlefield.Remove(necro); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	trace := walkSteps(t, g, 200, func() bool {
		return g.Turn.ActiveSeat == 2 && g.Turn.Step == game.StepPrecombatMain
	})
	if !sawStep(trace, 1, game.StepDraw) {
		t.Errorf("draw step still skipped after Necropotence left the battlefield")
	}
}

// --- Necropotence's ability ---------------------------------------

// TestNecropotenceExilesFaceDownThenDeliversAtYourEndStep is the
// sprint's second exit criterion, end to end.
func TestNecropotenceExilesFaceDownThenDeliversAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)
	needle := pushLibraryTopForTest(me, "Needle")

	lifeBefore, handBefore := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The life is a COST — paid at announce, while the ability is
	// still on the stack and nothing has been exiled.
	if me.Life != lifeBefore-1 {
		t.Errorf("life %d -> %d, want -1 paid at announce", lifeBefore, me.Life)
	}
	if !me.Library.Contains(needle) {
		t.Errorf("the card left the library before the ability resolved")
	}
	passPriorityAroundTable(t, g)

	exiled, ok := findInZone(g.Exile, needle)
	if !ok {
		t.Fatalf("top card was not exiled")
	}
	if !exiled.FaceDown {
		t.Errorf("Necropotence's exile must be face down")
	}
	// The assertion this card exists for: exile is a public zone and
	// every other exile in the engine marks all four seats knowers.
	// This one marks nobody, the controller included.
	for i, p := range g.Seats {
		if exiled.IsKnownTo(p.ID) {
			t.Errorf("seat %d can read a face-down exiled card", i)
		}
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand changed by %d on the exile; the card arrives at the end step",
			me.Hand.Size()-handBefore)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("want one delayed trigger queued, got %d", len(g.DelayedTriggers))
	}

	// Walk to this seat's own end step. The delayed trigger drains
	// on step entry and goes on the stack; priority passes resolve
	// it.
	walkSteps(t, g, 40, func() bool {
		return g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepEnd
	})
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(needle) {
		t.Fatalf("the exiled card never reached hand at the end step")
	}
	inHand, _ := findInZone(me.Hand, needle)
	if inHand.FaceDown {
		t.Errorf("a card in hand is not face down (CR 400.7)")
	}
	if !inHand.IsKnownTo(me.ID) {
		t.Errorf("the controller should be able to read the card once it is in hand")
	}
	if inHand.IsKnownTo(g.Seats[1].ID) {
		t.Errorf("an opponent can read a card in someone else's hand")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("delayed trigger queue not drained: %d left", len(g.DelayedTriggers))
	}
}

// TestNecropotenceHoldsTheCardUntilYOUREndStep — the printed text
// says "your next end step", so an opponent's end step in between
// must not deliver. ControllerTurnOnly on the delayed trigger is the
// one line that makes the difference, and without a test it reads
// like boilerplate.
func TestNecropotenceHoldsTheCardUntilYOUREndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)
	needle := pushLibraryTopForTest(me, "Needle")

	// Activate on seat 0's turn — Necropotence has no timing
	// restriction, so its controller may use it on anyone's turn.
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := findInZone(g.Exile, needle); !ok {
		t.Fatalf("top card was not exiled")
	}

	// Seat 0's end step comes first and is not "your" end step.
	walkSteps(t, g, 40, func() bool {
		return g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepEnd
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(needle) {
		t.Fatalf("the card was delivered at an OPPONENT's end step")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("the trigger should still be queued, have %d", len(g.DelayedTriggers))
	}

	walkSteps(t, g, 40, func() bool {
		return g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepEnd
	})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(needle) {
		t.Errorf("the card never arrived at its controller's own end step")
	}
}

// TestNecropotenceStacksMultipleActivations — five activations in
// one turn are five delayed triggers, and all five cards arrive
// together.
func TestNecropotenceStacksMultipleActivations(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)

	needles := make([]uuid.UUID, 0, 5)
	for i := 0; i < 5; i++ {
		needles = append(needles, pushLibraryTopForTest(me, "Needle"))
	}
	lifeBefore, handBefore := me.Life, me.Hand.Size()
	for i := 0; i < 5; i++ {
		if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	if me.Life != lifeBefore-5 {
		t.Errorf("life %d -> %d, want -5", lifeBefore, me.Life)
	}
	if len(g.DelayedTriggers) != 5 {
		t.Fatalf("want 5 delayed triggers, got %d", len(g.DelayedTriggers))
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand changed by %d before the end step", me.Hand.Size()-handBefore)
	}

	walkSteps(t, g, 40, func() bool {
		return g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepEnd
	})
	passPriorityAroundTable(t, g)
	for i, id := range needles {
		if !me.Hand.Contains(id) {
			t.Errorf("card %d never reached hand", i)
		}
	}
	if me.Hand.Size() != handBefore+5 {
		t.Errorf("hand delta %d, want +5", me.Hand.Size()-handBefore)
	}
}

// TestNecropotenceEmptyLibraryCostsLifeAndDoesNotKill — the ability
// is not a draw, so running the library out to it is not the
// CR 704.5b loss.
func TestNecropotenceEmptyLibraryCostsLifeAndDoesNotKill(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	necro := seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)
	me.Library.Cards = nil

	lifeBefore := me.Life
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Life != lifeBefore-1 {
		t.Errorf("life %d -> %d, want -1: the cost is paid regardless", lifeBefore, me.Life)
	}
	if me.LosesAtNextSBA {
		t.Errorf("exiling from an empty library is not drawing from an empty library")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("nothing was exiled, so nothing should be scheduled: %d queued",
			len(g.DelayedTriggers))
	}
}

// TestNecropotenceExilesDiscardedCards — the second clause. The card
// really reaches the graveyard first (it is a trigger, not a
// replacement); the exile happens when the trigger resolves.
func TestNecropotenceExilesDiscardedCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_ = seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)

	victim := soleHandCardForTest(me, "Chaff")
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	if !me.Graveyard.Contains(victim) {
		t.Fatalf("the discard should put the card in the graveyard first")
	}
	passPriorityAroundTable(t, g)

	if me.Graveyard.Contains(victim) {
		t.Errorf("the discarded card is still in the graveyard")
	}
	if !g.Exile.Contains(victim) {
		t.Errorf("the discarded card was not exiled")
	}
}

// TestNecropotenceIgnoresOpponentDiscards — "whenever YOU discard".
func TestNecropotenceIgnoresOpponentDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = seedReplacementPermanent(g, necropotenceOracle, "Necropotence", me.ID)

	victim := soleHandCardForTest(opp, "Chaff")
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(opp.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(victim) {
		t.Errorf("an opponent's discard was exiled by someone else's Necropotence")
	}
}

// --- Yawgmoth's Bargain / Griselbrand -----------------------------

func TestYawgmothsBargainDrawsForOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bargain := seedReplacementPermanent(g, yawgmothsBargainOracle, "Yawgmoth's Bargain", me.ID)

	lifeBefore, handBefore := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, bargain, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != lifeBefore-1 {
		t.Errorf("life %d -> %d, want -1 at announce", lifeBefore, me.Life)
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("the draw should wait for resolution")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand delta %d, want +1", me.Hand.Size()-handBefore)
	}
}

func TestGriselbrandDrawsSevenForSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gris := pushCatalogPermanent(g, me.ID, "Griselbrand",
		"Legendary Creature — Demon", griselbrandOracle, false)

	lifeBefore, handBefore := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != lifeBefore-7 {
		t.Errorf("life %d -> %d, want -7", lifeBefore, me.Life)
	}
	if me.Hand.Size() != handBefore+7 {
		t.Errorf("hand delta %d, want +7", me.Hand.Size()-handBefore)
	}
}

// TestGriselbrandCannotPayLifeItDoesNotHave — CR 119.4. The whole
// cost is validated before any of it is paid, so a failed activation
// leaves the life total alone.
func TestGriselbrandCannotPayLifeItDoesNotHave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gris := pushCatalogPermanent(g, me.ID, "Griselbrand",
		"Legendary Creature — Demon", griselbrandOracle, false)
	me.Life = 6

	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{}); err == nil {
		t.Errorf("activating at 6 life should be illegal")
	}
	if me.Life != 6 {
		t.Errorf("life changed to %d on a rejected activation", me.Life)
	}
}

// TestGriselbrandKeywords — flying and lifelink both ride
// PrintedKeywords; neither is decorative on this card.
func TestGriselbrandKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Griselbrand",
		TypeLine:   "Legendary Creature — Demon",
		OracleID:   griselbrandOracle,
		Power:      7, Toughness: 7,
		Owner: me.ID, Controller: me.ID,
	})
	for _, kw := range []string{"flying", "lifelink"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Griselbrand should have %s on the battlefield", kw)
		}
	}
}

// --- Vilis --------------------------------------------------------

// TestVilisDrawsForLifePaidToAnotherEngine — the payoff reads a life
// PAYMENT, not just a "lose life" effect, which is what makes it the
// card that pairs with the rest of this file. One activation of
// Yawgmoth's Bargain is one life lost, so Vilis draws one on top of
// the Bargain's own.
func TestVilisDrawsForLifePaidToAnotherEngine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_ = pushCatalogPermanent(g, me.ID, "Vilis, Broker of Blood",
		"Legendary Creature — Demon", vilisOracle, false)
	bargain := seedReplacementPermanent(g, yawgmothsBargainOracle, "Yawgmoth's Bargain", me.ID)

	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, bargain, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand delta %d, want +2 (one from the Bargain, one from Vilis)", got)
	}
}

// TestVilisDrawsOncePerLossNotPerPoint — losing 7 in one go is one
// trigger that draws seven, not seven triggers.
func TestVilisDrawsOncePerLossNotPerPoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_ = pushCatalogPermanent(g, me.ID, "Vilis, Broker of Blood",
		"Legendary Creature — Demon", vilisOracle, false)
	// Deep enough library that drawing seven cannot bottom out.
	for i := 0; i < 20; i++ {
		pushLibraryTopForTest(me, "Filler")
	}

	handBefore := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -7); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 7 {
		t.Errorf("hand delta %d, want +7 (one trigger for a 7-life loss)", got)
	}
}

// TestVilisIgnoresLifeGainAndOpponentLosses — the sign test and the
// seat test, the two ways this trigger goes wrong.
func TestVilisIgnoresLifeGainAndOpponentLosses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = pushCatalogPermanent(g, me.ID, "Vilis, Broker of Blood",
		"Legendary Creature — Demon", vilisOracle, false)

	handBefore := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -5)
	})
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("hand delta %d, want 0: gaining life and an opponent's loss are not triggers", got)
	}
}

// TestVilisShrinksATargetCreature — the -1/-1 half, and the ordering
// that comes with it: the 2 life is paid at announce, so the draw
// trigger goes on the stack above the shrink and resolves first.
func TestVilisShrinksATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vilis := pushCatalogPermanent(g, me.ID, "Vilis, Broker of Blood",
		"Legendary Creature — Demon", vilisOracle, false)
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Bears",
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})

	lifeBefore := me.Life
	if err := g.ActivateCatalogAbility(me.ID, vilis, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Life != lifeBefore-2 {
		t.Errorf("life %d -> %d, want -2", lifeBefore, me.Life)
	}
	if got := effectivePower(t, g, victim); got != 1 {
		t.Errorf("power %d after -1/-1, want 1", got)
	}
	if got := effectiveToughness(t, g, victim); got != 1 {
		t.Errorf("toughness %d after -1/-1, want 1", got)
	}
}

// --- Bloodgift Demon ----------------------------------------------

// TestBloodgiftDemonCanGiftAnOpponent — "target player", so the draw
// and the life loss both land on whoever was chosen, not on the
// controller.
func TestBloodgiftDemonCanGiftAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, gifted := g.Seats[1], g.Seats[2]
	demon := pushPermanentForTest(g, me.ID, "Bloodgift Demon",
		bloodgiftDemonOracle, "Creature — Demon")

	advanceToUpkeepOf(t, g, 1)
	// The trigger is targeted, so it waits on a pick_target prompt
	// before it reaches the stack at all (CR 603.3d).
	pickPlayer(t, g, me.ID, gifted.ID)
	if triggerOnStack(g, demon) == nil {
		t.Fatalf("Bloodgift Demon's upkeep trigger is not on the stack after targeting")
	}

	lifeBefore, handBefore := gifted.Life, gifted.Hand.Size()
	myLife, myHand := me.Life, me.Hand.Size()
	passPriorityAroundTable(t, g)

	if gifted.Hand.Size() != handBefore+1 {
		t.Errorf("target hand delta %d, want +1", gifted.Hand.Size()-handBefore)
	}
	if gifted.Life != lifeBefore-1 {
		t.Errorf("target life %d -> %d, want -1", lifeBefore, gifted.Life)
	}
	if me.Life != myLife || me.Hand.Size() != myHand {
		t.Errorf("the controller should be untouched when the gift goes elsewhere")
	}
}
