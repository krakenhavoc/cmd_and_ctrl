package game

import (
	"testing"

	"github.com/google/uuid"
)

// skip_this_turn_test.go — #6226: skips by choice. Fatespinner's "skips
// each instance of the chosen step or phase this turn", Time Vault's "If
// you would begin your turn while this artifact is tapped, you may skip
// that turn instead. If you do, untap this artifact", and Fasting's "If
// you would begin your draw step, you may skip that step instead. If you
// do, you gain 2 life" (CR 614.10, 614.10a, 500.11).

// walkTurnOf records every step of the turn in progress, from the cursor
// to its cleanup.
func walkThisTurn(t *testing.T, g *Game) []Step {
	t.Helper()
	seat, seq := g.Turn.ActiveSeat, g.Turn.Seq
	var steps []Step
	for i := 0; i < 40; i++ {
		if g.Turn.ActiveSeat != seat || g.Turn.Seq != seq {
			return steps
		}
		steps = append(steps, g.Turn.Step)
		if len(g.PendingChoices) > 0 {
			t.Fatalf("a prompt opened at %s", g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("the turn never ended")
	return nil
}

func stepsHave(steps []Step, want func(Step) bool) bool {
	for _, s := range steps {
		if want(s) {
			return true
		}
	}
	return false
}

func isMainStep(s Step) bool { return s == StepPrecombatMain || s == StepPostcombatMain }
func isCombatStep(s Step) bool {
	return PhaseOf(s) == PhaseCombat
}

// Fatespinner, main phase: both main phases of that turn are skipped
// (CR 500.11), and the next turn has its main phases again.
func TestSkipThisTurnMainSkipsBothMainPhases(t *testing.T) {
	g := newActiveGame(t) // seat 0's upkeep
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		if !g.SkipThisTurnForEffect(me, uuid.Nil, SkipThisTurnMain, "Fatespinner") {
			t.Fatal("no record written")
		}
	})
	steps := walkThisTurn(t, g)
	if stepsHave(steps, isMainStep) {
		t.Fatalf("a main phase began: %v", steps)
	}
	if !stepsHave(steps, isCombatStep) || !stepsHave(steps, func(s Step) bool { return s == StepEnd }) {
		t.Errorf("the rest of the turn should happen: %v", steps)
	}
	next := walkThisTurn(t, g)
	if !stepsHave(next, func(s Step) bool { return s == StepPrecombatMain }) {
		t.Errorf("the next turn lost its main phase: %v", next)
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("the record outlived its turn: %+v", g.ScopedEffects)
	}
}

// Fatespinner, combat phase and draw step.
func TestSkipThisTurnCombatAndDraw(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		g.SkipThisTurnForEffect(me, uuid.Nil, SkipThisTurnCombat, "Fatespinner")
		g.SkipThisTurnForEffect(me, uuid.Nil, SkipThisTurnDraw, "Fatespinner")
	})
	hand := g.Seats[0].Hand.Size()
	steps := walkThisTurn(t, g)
	if stepsHave(steps, isCombatStep) || stepsHave(steps, func(s Step) bool { return s == StepDraw }) {
		t.Fatalf("a skipped step began: %v", steps)
	}
	if !stepsHave(steps, isMainStep) {
		t.Errorf("the main phases should happen: %v", steps)
	}
	if g.Seats[0].Hand.Size() != hand {
		t.Errorf("a card was drawn in a skipped draw step")
	}
}

// Another player's record does not skip the active player's phases.
func TestSkipThisTurnBindsOnlyItsPlayer(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1].ID
	g.WithWriteLock(func() { g.SkipThisTurnForEffect(opp, uuid.Nil, SkipThisTurnMain, "Fatespinner") })
	if !stepsHave(walkThisTurn(t, g), isMainStep) {
		t.Error("the active player's main phases were skipped by another player's record")
	}
}

func TestSkipThisTurnModProblem(t *testing.T) {
	p := uuid.New()
	for _, m := range []Mod{
		{Kind: ModSkipThisTurn, Amount: 1, Text: SkipThisTurnMain},
		{Kind: ModSkipThisTurn, Player: p, Text: SkipThisTurnMain},
		{Kind: ModSkipThisTurn, Player: p, Amount: 1, Text: "upkeep"},
	} {
		if skipThisTurnModProblem(m) == "" {
			t.Errorf("%+v was accepted", m)
		}
	}
	if pr := skipThisTurnModProblem(Mod{Kind: ModSkipThisTurn, Player: p, Amount: 3, Text: SkipThisTurnCombat}); pr != "" {
		t.Errorf("a sound record was refused: %s", pr)
	}
}

func withCatalogTurnSkipOffers(t *testing.T, key string) {
	t.Helper()
	prev := CatalogTurnSkipOffers
	CatalogTurnSkipOffers = func(k string) []TurnSkipOffer {
		if k != key {
			return nil
		}
		return []TurnSkipOffer{{WhileTapped: true, UntapSource: true, Label: "Time Vault"}}
	}
	t.Cleanup(func() { CatalogTurnSkipOffers = prev })
}

func openTurnSkipOffer(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceConfirm && c.PickThen == turnSkipOfferThen.Key() {
			return c
		}
	}
	t.Fatal("no turn-skip offer is open")
	return nil
}

// Time Vault, tapped: as its controller would begin a turn, they may
// skip it instead, and it untaps. Nothing of the skipped turn begins and
// the next player's turn follows.
func TestTimeVaultSkipsTheTurnAndUntaps(t *testing.T) {
	g := newActiveGame(t) // seat 0's turn
	opp := g.Seats[1].ID
	const vault = "time-vault"
	v := pushTappedPermanent(g, opp, "Time Vault", vault, "Artifact", true)
	withCatalogTurnSkipOffers(t, vault)
	seq := g.Turn.Seq

	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	offer := openTurnSkipOffer(t, g)
	if offer.Chooser != opp {
		t.Fatalf("chooser = %v, want the vault's controller", offer.Chooser)
	}
	if !g.TurnChangePending || g.Turn.Seq != seq {
		t.Fatal("a turn began before the offer was answered")
	}

	r := restoreStrictForTest(t, g)
	if err := r.ResolveConfirm(offer.ID, opp, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if r.TurnChangePending {
		t.Fatal("the turn change did not resume")
	}
	if r.Turn.ActiveSeat != 0 {
		t.Errorf("active seat = %d, want seat 0 again (seat 1's turn skipped)", r.Turn.ActiveSeat)
	}
	if cardByIDForUntapTest(r, v).Tapped {
		t.Error("Time Vault did not untap")
	}
	if r.TurnSkipsForLocked(opp) != 0 {
		t.Error("the skip charge was not spent on that turn")
	}
}

// Declining takes the turn; an untapped Time Vault asks nothing.
func TestTimeVaultDeclinedTakesTheTurn(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1].ID
	const vault = "time-vault"
	v := pushTappedPermanent(g, opp, "Time Vault", vault, "Artifact", true)
	withCatalogTurnSkipOffers(t, vault)

	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	offer := openTurnSkipOffer(t, g)
	if err := g.ResolveConfirm(offer.ID, opp, false); err != nil {
		t.Fatal(err)
	}
	if g.TurnChangePending || g.Turn.ActiveSeat != 1 {
		t.Fatalf("seat 1's turn did not begin (pending=%v, seat=%d)", g.TurnChangePending, g.Turn.ActiveSeat)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the declined offer was asked again: %d prompts", len(g.PendingChoices))
	}
	// Time Vault doesn't untap during its controller's untap step in the
	// real card (a separate restriction); here nothing holds it, so the
	// untap step untapped it, and an untapped vault offers nothing.
	if cardByIDForUntapTest(g, v).Tapped {
		t.Skip("the vault stayed tapped; the untapped case is not reachable here")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("an untapped Time Vault offered a skip")
	}
}

// The offer applies to an extra turn too: "If you would begin your turn".
func TestTimeVaultOffersOnAnExtraTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	const vault = "time-vault"
	pushTappedPermanent(g, me, "Time Vault", vault, "Artifact", true)
	withCatalogTurnSkipOffers(t, vault)
	g.WithWriteLock(func() { g.TakeExtraTurnsForEffect(me, uuid.Nil, 1) })

	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	offer := openTurnSkipOffer(t, g)
	if err := g.ResolveConfirm(offer.ID, me, true); err != nil {
		t.Fatal(err)
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Extra {
		t.Errorf("the extra turn was not skipped (seat=%d, extra=%v)", g.Turn.ActiveSeat, g.Turn.Extra)
	}
}

// Fasting: "If you would begin your draw step, you may skip that step
// instead. If you do, you gain 2 life." An Optional step-transition
// replacement whose Replace cancels and gains — the shape develop
// already has (ReplacementEffect.Optional); pinned here.
func TestOptionalDrawStepSkipGainsLife(t *testing.T) {
	fasting := ReplacementEffect{
		Watches: []EventKind{EventStepTransition},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventStepTransition && ev.StepTransitionStep == StepDraw && ev.StepTransitionSeat == 1
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			ev.Cancel()
			return g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, 2)
		},
		Optional: true,
		Label:    "Fasting",
	}
	for _, accept := range []bool{true, false} {
		g := newGameWithStepReplacements(t, fasting)
		life, hand := g.Seats[1].Life, g.Seats[1].Hand.Size()
		advanceUntil(t, g, 40, func() bool { return g.Turn.ActiveSeat == 1 && g.Turn.Step == StepPrecombatMain })
		if len(g.PendingChoices) != 1 {
			t.Fatalf("accept=%v: prompts = %d, want the may", accept, len(g.PendingChoices))
		}
		c := g.PendingChoices[0]
		if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, accept); err != nil {
			t.Fatal(err)
		}
		if accept {
			if g.Turn.Step == StepDraw || g.Seats[1].Hand.Size() != hand || g.Seats[1].Life != life+2 {
				t.Errorf("skip: step=%s hand %d→%d life %d→%d", g.Turn.Step, hand, g.Seats[1].Hand.Size(), life, g.Seats[1].Life)
			}
		} else if g.Seats[1].Hand.Size() != hand+1 || g.Seats[1].Life != life {
			t.Errorf("decline: hand %d→%d life %d→%d", hand, g.Seats[1].Hand.Size(), life, g.Seats[1].Life)
		}
	}
}
