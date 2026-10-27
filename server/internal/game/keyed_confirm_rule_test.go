package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// keyed_confirm_rule_test.go — #4264: a keyed confirm, a keyed
// choose_cards pick's set rule and its string carry survive a restore,
// and each registry refuses a key this binary lacks.

// --- the keyed confirm -----------------------------------------------

var (
	keyedConfirmSeen []ConfirmAnswer
	keyedConfirm     = RegisterConfirmThen(testEffectKeyPrefix+"keyed-confirm", func(_ *Game, r ConfirmAnswer) error {
		keyedConfirmSeen = append(keyedConfirmSeen, r)
		return nil
	})
)

func queueKeyedConfirm(t *testing.T, g *Game, chooser uuid.UUID, carry uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.QueueConfirmKeyedForEffect(ConfirmPrompt{
			Chooser:  chooser,
			Source:   uuid.New(),
			Question: "draw a card?",
		}, keyedConfirm, []uuid.UUID{carry})
		if err != nil {
			t.Fatal(err)
		}
	})
	return id
}

func restoreStrictForTest(t *testing.T, g *Game) *Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the table is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	return restored
}

func TestKeyedConfirmSurvivesARestoreAndAnswersBothWays(t *testing.T) {
	for _, accept := range []bool{true, false} {
		keyedConfirmSeen = nil
		g := newActiveGame(t)
		me := g.Seats[0]
		carried := uuid.New()
		id := queueKeyedConfirm(t, g, me.ID, carried)
		restored := restoreStrictForTest(t, g)
		if err := restored.ResolveConfirm(id, me.ID, accept); err != nil {
			t.Fatalf("ResolveConfirm on the restored game: %v", err)
		}
		if len(keyedConfirmSeen) != 1 {
			t.Fatalf("accept=%v: the answer ran %d times, want 1", accept, len(keyedConfirmSeen))
		}
		r := keyedConfirmSeen[0]
		if r.Accepted != accept || r.Chooser != me.ID || len(r.Carry) != 1 || r.Carry[0] != carried {
			t.Errorf("accept=%v: answer got %+v", accept, r)
		}
		if len(restored.PendingChoices) != 0 {
			t.Errorf("the answered confirm is still open: %+v", restored.PendingChoices)
		}
	}
}

// A keyed confirm is a plain "you may" with no labels, so a standing
// answer covers it, as it does a closure confirm.
func TestKeyedConfirmTakesAStandingAnswer(t *testing.T) {
	g := newActiveGame(t)
	id := queueKeyedConfirm(t, g, g.Seats[0].ID, uuid.New())
	g.WithWriteLock(func() {
		_, c := g.findChoiceLocked(id)
		if !confirmCovered(c) {
			t.Error("a keyed confirm is not covered by a standing answer")
		}
	})
}

func TestKeyedConfirmRefusesAnUnknownKeyAtRestore(t *testing.T) {
	g := newActiveGame(t)
	queueKeyedConfirm(t, g, g.Seats[0].ID, uuid.New())
	snap := g.CaptureSnapshot()
	for i := range snap.PendingChoices {
		snap.PendingChoices[i].PickThen = "test/never-registered"
	}
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown confirm answer: err = %v, want ErrUnknownEffectKey", err)
	}
}

// --- a keyed choose_cards pick's set rule and word --------------------

var (
	keyedRuleSeen []CardsChosen
	keyedRule     = RegisterChooseCardsThenRule(testEffectKeyPrefix+"keyed-set-rule",
		func(picked []Card) bool { return len(picked) == 2 && picked[0].Name != picked[1].Name },
		func(_ *Game, r CardsChosen) error {
			keyedRuleSeen = append(keyedRuleSeen, r)
			return nil
		})
)

func TestKeyedChooseCardsKeepsItsSetRuleAndWordAcrossARestore(t *testing.T) {
	keyedRuleSeen = nil
	g := newActiveGame(t)
	me := g.Seats[0]
	var ids []uuid.UUID
	for _, name := range []string{"Alpha", "Alpha", "Beta"} {
		c := NewCard(name, me.ID)
		me.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:   me.ID,
			Source:    uuid.New(),
			Question:  "two different names",
			Cards:     ids,
			Min:       0,
			Max:       2,
			ThenKey:   keyedRule,
			CarryWord: "Eldrazi",
		})
	})
	restored := restoreStrictForTest(t, g)
	if err := restored.ResolveChooseCards(id, me.ID, ids[:1]); !errors.Is(err, ErrChoiceSetRejected) {
		t.Errorf("one card: err = %v, want ErrChoiceSetRejected", err)
	}
	if err := restored.ResolveChooseCards(id, me.ID, ids[:2]); !errors.Is(err, ErrChoiceSetRejected) {
		t.Errorf("two copies: err = %v, want ErrChoiceSetRejected", err)
	}
	var legal bool
	restored.WithWriteLock(func() {
		_, c := restored.findChoiceLocked(id)
		legal = restored.ChooseCardsPickLegalLocked(c, []uuid.UUID{ids[0], ids[2]})
	})
	if !legal {
		t.Error("the enumerator's window refuses a legal pair")
	}
	if len(keyedRuleSeen) != 0 {
		t.Fatalf("the continuation ran on a refused answer: %+v", keyedRuleSeen)
	}
	if err := restored.ResolveChooseCards(id, me.ID, nil); err != nil {
		t.Fatalf("choosing nothing is always allowed: %v", err)
	}
	if len(keyedRuleSeen) != 1 || keyedRuleSeen[0].Word != "Eldrazi" {
		t.Fatalf("continuation got %+v, want one run with the word", keyedRuleSeen)
	}
}
