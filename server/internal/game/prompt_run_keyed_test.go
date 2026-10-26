package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// prompt_run_keyed_test.go pins #4290: a prompted run whose continuation
// is a registered key is plain data, written to a restore point with its
// legs, so a table waiting on any leg of a multi-seat discard, a
// fire-and-forget discard fan-out, or a keyed sacrifice run restores
// strictly and finishes on the restored table.
//
// Each restore test answers the first leg, takes a STRICT restore through
// JSON with the second still open, and answers on the restored table.

// keyedRunCall records the keyed continuations' calls. Registered
// functions must capture nothing, so the record is package state; each
// test resets it.
var keyedRunCall struct {
	ran  int
	got  RunDone
	legs []RunLeg
}

func resetKeyedRunCall() {
	keyedRunCall.ran = 0
	keyedRunCall.got = RunDone{}
	keyedRunCall.legs = nil
}

func recordKeyedRun(_ *Game, r RunDone) error {
	keyedRunCall.ran++
	keyedRunCall.got = r
	return nil
}

var (
	testKeyedRun    = RegisterPromptRunThen(testEffectKeyPrefix+"keyed-run", recordKeyedRun)
	testKeyedRunLeg = RegisterPromptRunThenLeg(testEffectKeyPrefix+"keyed-run-leg",
		func(_ *Game, leg RunLeg) error {
			keyedRunCall.legs = append(keyedRunCall.legs, leg)
			return nil
		}, recordKeyedRun)
)

// strictRestore takes g through capture → JSON → RestoreStrict and fails
// the test when the table is not a restore point.
func strictRestore(t *testing.T, g *Game) *Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the table is not a restore point: %+v", snap.Continuations)
	}
	restored, err := throughJSON(t, snap).RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}

// keyedDiscardRunTable is three seats with one card each asked by a keyed
// fan-out ("each other player discards a card"), with seat 3 the caster
// on the carry. It returns the cards dealt and the source.
func keyedDiscardRunTable(t *testing.T, key PromptRunThen) (*Game, map[uuid.UUID]uuid.UUID, uuid.UUID) {
	t.Helper()
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[3]
	source := departureTestSource(g, caster.ID, "Syphon Mind")
	pitched := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats[:3] {
		pitched[p.ID] = oneCardHand(t, g, p, "Pitch")
	}
	g.WithWriteLock(func() {
		if err := g.EachPlayerDiscardsThenKeyedForEffect(caster.ID, DiscardPrompt{Source: source, N: 1},
			key, []uuid.UUID{caster.ID}); err != nil {
			t.Fatalf("EachPlayerDiscardsThenKeyedForEffect: %v", err)
		}
	})
	if openDiscardPrompts(g) != 3 {
		t.Fatalf("expected three prompts, got %d", openDiscardPrompts(g))
	}
	return g, pitched, source
}

// TestAKeyedDiscardRunIsARestorePointWithALegAnswered is the headline:
// one seat has answered, two are still open, and the table restores
// strictly. The restored run still holds the first seat's discard, and
// pays out once, after the last answer on the restored table.
func TestAKeyedDiscardRunIsARestorePointWithALegAnswered(t *testing.T) {
	g, pitched, source := keyedDiscardRunTable(t, testKeyedRun)
	seats := []uuid.UUID{g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID}
	caster := g.Seats[3].ID
	answerDiscard(t, g, seats[0], pitched[seats[0]])
	if keyedRunCall.ran != 0 {
		t.Fatal("the run paid out with two legs open")
	}

	snap := g.CaptureSnapshot()
	if len(snap.PromptRuns) != 1 || snap.PromptRuns[0].Outstanding != 2 ||
		len(snap.PromptRuns[0].Landed) != 1 || snap.PromptRuns[0].Landed[0].Seat != seats[0] {
		t.Fatalf("the snapshot's run is %+v, want one run with two legs outstanding and the first seat's discard", snap.PromptRuns)
	}
	for _, c := range snap.PendingChoices {
		if c.PromptRun != snap.PromptRuns[0].ID || c.PickThen != discardLegThen.Key() {
			t.Errorf("a leg on disk is %+v, want the run link and the engine's discard key", c)
		}
	}

	g = strictRestore(t, g)
	answerDiscard(t, g, seats[1], pitched[seats[1]])
	if keyedRunCall.ran != 0 {
		t.Fatal("the restored run paid out with a leg still open")
	}
	answerDiscard(t, g, seats[2], pitched[seats[2]])
	if keyedRunCall.ran != 1 {
		t.Fatalf("the restored run ran %d times, want 1", keyedRunCall.ran)
	}
	got := keyedRunCall.got
	if got.Source != source || len(got.Carry) != 1 || got.Carry[0] != caster {
		t.Errorf("the continuation was handed source %v carry %v, want %v and the caster", got.Source, got.Carry, source)
	}
	if got.Discards().Count() != 3 {
		t.Errorf("the run discarded %d cards, want 3 (%v)", got.Discards().Count(), got.Landed)
	}
	for i, e := range got.Landed {
		if e.Seat != seats[i] || len(e.Cards) != 1 || e.Cards[0] != pitched[e.Seat] {
			t.Errorf("entry %d is %+v, want seat %s with its card, in ask order", i, e, seats[i])
		}
		if !g.playerByIDLocked(e.Seat).Graveyard.Contains(pitched[e.Seat]) {
			t.Errorf("seat %s's card is not in its graveyard", e.Seat)
		}
	}
	if g.promptRuns != nil {
		t.Errorf("a finished run left %d runs behind", len(g.promptRuns))
	}
}

// TestAClosureRunIsStillCountedOnce: the closure form is not a restore
// point, and the census counts the run once rather than once per leg.
func TestAClosureRunIsStillCountedOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	source := departureTestSource(g, g.Seats[3].ID, "Syphon Mind")
	for _, p := range g.Seats[:3] {
		oneCardHand(t, g, p, "Pitch")
	}
	eachPlayerDiscardsThen(t, g, source, g.Seats[3].ID, 1)
	snap := g.CaptureSnapshot()
	if snap.Restorable() {
		t.Fatal("a closure run is a restore point")
	}
	if len(snap.PromptRuns) != 0 {
		t.Errorf("a closure run was written: %+v", snap.PromptRuns)
	}
	// The legs are closure prompts too (each its own frame), so count
	// the run's share alone: one note naming the run.
	runs := 0
	for _, l := range snap.Continuations.Labels {
		if l == "a prompted run waits on 3 prompt(s) with a continuation closure" {
			runs++
		}
	}
	if runs != 1 {
		t.Errorf("the census names the run %d times, want once: %v", runs, snap.Continuations.Labels)
	}
}

// TestAKeyedRunWithAnUnknownKeyIsRefused is the rollback case: a binary
// that does not know the run's key, or the engine's discard-leg key,
// refuses the file rather than restoring legs that settle into nothing.
func TestAKeyedRunWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _, _ := keyedDiscardRunTable(t, testKeyedRun)
	snap := throughJSON(t, g.CaptureSnapshot())
	snap.PromptRuns[0].Key = testEffectKeyPrefix + "never-registered"
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown run key: err = %v, want ErrUnknownEffectKey", err)
	}
	snap = throughJSON(t, g.CaptureSnapshot())
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = "choose-cards/never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown leg key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// TestASeatLeavingSettlesItsKeyedDiscardLeg: departure (dropDefault)
// settles a keyed leg with nothing discarded, on a restored table too,
// and the run still pays out after the last answer.
func TestASeatLeavingSettlesItsKeyedDiscardLeg(t *testing.T) {
	g, pitched, _ := keyedDiscardRunTable(t, testKeyedRun)
	first, leaver, last := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	answerDiscard(t, g, first, pitched[first])
	g = strictRestore(t, g)
	if err := g.Concede(leaver); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if keyedRunCall.ran != 0 {
		t.Fatal("the run paid out with the last seat's leg still open")
	}
	answerDiscard(t, g, last, pitched[last])
	if keyedRunCall.ran != 1 {
		t.Fatalf("the run ran %d times, want 1 — a departure must not strand it", keyedRunCall.ran)
	}
	d := keyedRunCall.got.Discards()
	if d.Discarded(leaver) || !d.Discarded(first) || !d.Discarded(last) || d.Count() != 2 {
		t.Errorf("the run's answer is %v, want the two seats that answered and nothing for the one that left", d)
	}
}

// TestAKeyedDiscardRunsItsPerSeatTail: the per-seat tail runs for each
// seat as its discard lands, before the run settles, and at once for a
// seat whose hand was empty (CR 701.9a), as the closure form's leg Then
// does.
func TestAKeyedDiscardRunsItsPerSeatTail(t *testing.T) {
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	caster, empty, asked := g.Seats[3], g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Vicious Rumors")
	g.WithWriteLock(func() { empty.Hand.Cards = nil })
	pitch := oneCardHand(t, g, asked, "Pitch")
	g.WithWriteLock(func() { g.Seats[2].Hand.Cards = nil })
	g.WithWriteLock(func() {
		if err := g.EachPlayerDiscardsThenKeyedForEffect(caster.ID, DiscardPrompt{Source: source, N: 1},
			testKeyedRunLeg, []uuid.UUID{caster.ID}); err != nil {
			t.Fatalf("EachPlayerDiscardsThenKeyedForEffect: %v", err)
		}
	})
	if len(keyedRunCall.legs) != 2 || keyedRunCall.legs[0].Seat != empty.ID || keyedRunCall.legs[1].Seat != g.Seats[2].ID {
		t.Fatalf("the empty-handed seats' tails are %+v, want both at once, in ask order", keyedRunCall.legs)
	}
	g = strictRestore(t, g)
	answerDiscard(t, g, asked.ID, pitch)
	if len(keyedRunCall.legs) != 3 {
		t.Fatalf("tails %+v, want the answering seat's after its discard", keyedRunCall.legs)
	}
	leg := keyedRunCall.legs[2]
	if leg.Seat != asked.ID || len(leg.Landed) != 1 || leg.Landed[0] != pitch || leg.Source != source ||
		len(leg.Carry) != 1 || leg.Carry[0] != caster.ID {
		t.Errorf("the answering seat's tail was handed %+v", leg)
	}
	if keyedRunCall.ran != 1 || keyedRunCall.got.Discards().Count() != 1 {
		t.Errorf("the run ran %d times with %v, want once with the one discard", keyedRunCall.ran, keyedRunCall.got.Landed)
	}
}

// TestAFireAndForgetDiscardFanOutIsARestorePoint: "each opponent
// discards two cards" with nothing waiting (Tasigur's Cruelty) queues
// keyed legs with no run, so the table restores strictly with one seat
// answered and the rest discard on the restored table.
func TestAFireAndForgetDiscardFanOutIsARestorePoint(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[0]
	source := departureTestSource(g, caster.ID, "Tasigur's Cruelty")
	pitched := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats[1:] {
		pitched[p.ID] = oneCardHand(t, g, p, "Pitch")
	}
	var queued int
	g.WithWriteLock(func() { queued = g.EachPlayerDiscardsForEffect(caster.ID, DiscardPrompt{Source: source, N: 1}) })
	if queued != 3 {
		t.Fatalf("queued %d prompts, want 3", queued)
	}
	for _, c := range g.PendingChoices {
		if c.PickThen != discardLegThen.Key() || c.promptRun != uuid.Nil {
			t.Fatalf("a fire-and-forget leg is %+v, want a keyed leg with no run", c)
		}
	}
	first := g.Seats[1].ID
	answerDiscard(t, g, first, pitched[first])
	g = strictRestore(t, g)
	for _, p := range g.Seats[2:] {
		answerDiscard(t, g, p.ID, pitched[p.ID])
	}
	for _, p := range g.Seats[1:] {
		if !p.Graveyard.Contains(pitched[p.ID]) {
			t.Errorf("seat %s's card is not in its graveyard", p.Name)
		}
	}
	if g.promptRuns != nil {
		t.Errorf("a fan-out nobody waits on left %d runs behind", len(g.promptRuns))
	}
}

// keyedSacrificeTable is one seat owing two sacrifices to a keyed run
// ("sacrifice two creatures"), with its bears returned.
func keyedSacrificeTable(t *testing.T) (*Game, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	resetKeyedRunCall()
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Lotus Field")
	first, second := pushBear(g, me.ID), pushBear(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.PlayerSacrificesThenKeyedForEffect(source, me.ID, nil, "sacrifice a permanent", 2,
			testKeyedRun, []uuid.UUID{me.ID}); err != nil {
			t.Fatalf("PlayerSacrificesThenKeyedForEffect: %v", err)
		}
	})
	if openSacrificePrompts(g) != 2 {
		t.Fatalf("expected two prompts for one instruction, got %d", openSacrificePrompts(g))
	}
	return g, source, first, second
}

// TestAKeyedSacrificeRunIsARestorePointWithALegAnswered: a sacrifice
// prompt is data already, so with the run keyed the table restores
// strictly between the two answers of one instruction.
func TestAKeyedSacrificeRunIsARestorePointWithALegAnswered(t *testing.T) {
	g, source, first, second := keyedSacrificeTable(t)
	me := g.Seats[0].ID
	answerSacrifice(t, g, me, first)
	if keyedRunCall.ran != 0 {
		t.Fatal("the run paid out after the first of two answers")
	}
	g = strictRestore(t, g)
	if c := sacrificePromptFor(g, me); c == nil || c.PickThen != testKeyedRun.Key() || c.promptRun == uuid.Nil {
		t.Fatalf("the restored leg is %+v, want the run's key and link", c)
	}
	answerSacrifice(t, g, me, second)
	if keyedRunCall.ran != 1 {
		t.Fatalf("the restored run ran %d times, want 1", keyedRunCall.ran)
	}
	got := keyedRunCall.got
	if got.Source != source || got.Sacrifices().Count() != 2 || len(got.Landed) != 1 {
		t.Errorf("the run was handed %+v, want both bears under one entry and the source", got)
	}
}

// TestAKeyedSacrificeLegWithAnUnknownKeyIsRefused: the leg names its
// run's key, so a binary without it refuses the file even before it
// reads the run.
func TestAKeyedSacrificeLegWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _, _, _ := keyedSacrificeTable(t)
	snap := throughJSON(t, g.CaptureSnapshot())
	snap.PromptRuns = nil
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = testEffectKeyPrefix + "never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown sacrifice-leg key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// TestASeatLeavingSettlesItsKeyedSacrificeLegs: the sacrificing seat
// leaves with a leg open on the restored table, and the run still runs
// with what it had sacrificed.
func TestASeatLeavingSettlesItsKeyedSacrificeLegs(t *testing.T) {
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	victim := g.Seats[1]
	source := departureTestSource(g, g.Seats[0].ID, "Consuming Vapors")
	first, _ := pushBear(g, victim.ID), pushBear(g, victim.ID)
	g.WithWriteLock(func() {
		if err := g.PlayerSacrificesThenKeyedForEffect(source, victim.ID, nil, "sacrifice a permanent", 2,
			testKeyedRun, nil); err != nil {
			t.Fatalf("PlayerSacrificesThenKeyedForEffect: %v", err)
		}
	})
	answerSacrifice(t, g, victim.ID, first)
	g = strictRestore(t, g)
	if err := g.Concede(victim.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if keyedRunCall.ran != 1 {
		t.Fatalf("the run ran %d times, want 1 — a departure must not strand it", keyedRunCall.ran)
	}
	if by := keyedRunCall.got.Sacrifices().By(victim.ID); len(by) != 1 || by[0] != first {
		t.Errorf("the run's answer is %v, want the one bear sacrificed before the seat left", keyedRunCall.got.Landed)
	}
}

// TestAnnihilatorSacrificeChoiceIsARestorePoint: annihilator's "defending
// player sacrifices N permanents" is a keyed permanent-pick run, and the
// leg's frame (the battlefield re-check and the settle) is rebuilt from
// the prompt, so the table restores strictly with the choice open and the
// picks are sacrificed on the restored table.
func TestAnnihilatorSacrificeChoiceIsARestorePoint(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushAnnihilatorForTest(g, me.ID, "annihilator 2")
	fodder := pushFodderForTest(g, opp.ID, 3)
	attackWithAnnihilator(t, g, attacker, opp.ID)
	passUntilPrompt(t, g)
	if c := pendingChoiceOfKind(g, PendingChoiceOwnPermanents); c == nil || c.PickThen != annihilatorSacrificed.Key() {
		t.Fatalf("the annihilator prompt is %+v, want a keyed leg", c)
	}
	g = strictRestore(t, g)
	c := pendingChoiceOfKind(g, PendingChoiceOwnPermanents)
	if c == nil || c.chooseCardsResume == nil || c.chooseCardsResume.zone != ZoneBattlefield {
		t.Fatalf("the restored prompt is %+v, want its battlefield frame rebuilt", c)
	}
	if err := g.ResolveOwnPermanents(c.ID, opp.ID, fodder[:2]); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	for _, id := range fodder[:2] {
		if findBattlefieldCard(g, id) != nil {
			t.Errorf("%s was not sacrificed on the restored table", id)
		}
	}
	if findBattlefieldCard(g, fodder[2]) == nil {
		t.Error("the permanent not chosen was sacrificed")
	}
}

// TestAKeyedPermanentPickRunIsARestorePointWithALegAnswered: one chooser
// walks two boards (an own_permanents and a their_permanents leg); the
// first is answered, the table restores, and the run pays out with both
// after the second.
func TestAKeyedPermanentPickRunIsARestorePointWithALegAnswered(t *testing.T) {
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Tragic Arrogance")
	mine, theirs := pushFodderForTest(g, me.ID, 2), pushFodderForTest(g, opp.ID, 2)
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenKeyedForEffect(PermanentPickPrompt{
			Chooser: me.ID, Source: source, Question: "choose one",
			Of: []uuid.UUID{me.ID, opp.ID},
			Candidates: func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == of && c.Name == "Fodder" {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
		}, testKeyedRun, RunData{Carry: []uuid.UUID{opp.ID}}); err != nil {
			t.Fatalf("PermanentsPickedThenKeyedForEffect: %v", err)
		}
	})
	own := pendingChoiceOfKind(g, PendingChoiceOwnPermanents)
	if own == nil {
		t.Fatal("no own_permanents leg")
	}
	if err := g.ResolveOwnPermanents(own.ID, me.ID, mine[:1]); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	g = strictRestore(t, g)
	their := pendingChoiceOfKind(g, PendingChoiceTheirPermanents)
	if their == nil {
		t.Fatal("the restored table lost the their_permanents leg")
	}
	if err := g.ResolveTheirPermanents(their.ID, me.ID, theirs[:1]); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}
	if keyedRunCall.ran != 1 {
		t.Fatalf("the restored run ran %d times, want 1", keyedRunCall.ran)
	}
	picks := keyedRunCall.got.Picks()
	if by := picks.By(me.ID); len(by) != 1 || by[0] != mine[0] {
		t.Errorf("my board's pick is %v, want %v", by, mine[0])
	}
	if by := picks.By(opp.ID); len(by) != 1 || by[0] != theirs[0] {
		t.Errorf("their board's pick is %v, want %v", by, theirs[0])
	}
}

// TestAKeyedPermanentPickWithAValidateIsRefused: a set rule is a
// closure, so the keyed form will not take one.
func TestAKeyedPermanentPickWithAValidateIsRefused(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		_, err := g.PermanentsPickedThenKeyedForEffect(PermanentPickPrompt{
			Chooser: me.ID, Of: []uuid.UUID{me.ID},
			Validate: func([]Card) bool { return true },
		}, testKeyedRun, RunData{})
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("err = %v, want ErrInvalidParam", err)
		}
	})
}

// TestAKeyedPermanentPickLegWithAnUnknownKeyIsRefused: the leg names its
// run's key, so a binary without it refuses the file at the leg.
func TestAKeyedPermanentPickLegWithAnUnknownKeyIsRefused(t *testing.T) {
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	victim := g.Seats[1]
	pushFodderForTest(g, victim.ID, 2)
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenKeyedForEffect(PermanentPickPrompt{
			Chooser: victim.ID, Question: "sacrifice one",
			Of:         []uuid.UUID{victim.ID},
			Candidates: annihilatorCandidates(1),
		}, testKeyedRun, RunData{}); err != nil {
			t.Fatalf("PermanentsPickedThenKeyedForEffect: %v", err)
		}
	})
	snap := throughJSON(t, g.CaptureSnapshot())
	snap.PromptRuns = nil
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = testEffectKeyPrefix + "never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("restore with an unknown permanent-pick leg key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// TestAChooserLeavingSettlesTheirKeyedPermanentPick: departure settles an
// own_permanents leg of a keyed run with nothing picked, on a restored
// table.
func TestAChooserLeavingSettlesTheirKeyedPermanentPick(t *testing.T) {
	resetKeyedRunCall()
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	source := departureTestSource(g, g.Seats[0].ID, "Annihilator")
	pushFodderForTest(g, leaver.ID, 2)
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenKeyedForEffect(PermanentPickPrompt{
			Chooser: leaver.ID, Source: source, Question: "sacrifice one",
			Of:         []uuid.UUID{leaver.ID},
			Candidates: annihilatorCandidates(1),
		}, testKeyedRun, RunData{}); err != nil {
			t.Fatalf("PermanentsPickedThenKeyedForEffect: %v", err)
		}
	})
	g = strictRestore(t, g)
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if keyedRunCall.ran != 1 || keyedRunCall.got.Picks().Count() != 0 {
		t.Errorf("the run ran %d times with %v, want once with nothing picked", keyedRunCall.ran, keyedRunCall.got.Landed)
	}
}
