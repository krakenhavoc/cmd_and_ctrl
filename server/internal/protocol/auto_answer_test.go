package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_answer_test.go — ADR 0127 §6 and §8: the log line every seat
// reads, and the fields only the chooser gets.

func TestAutoAnswerLogLines(t *testing.T) {
	for _, c := range []struct {
		call, label, card, want string
	}{
		{game.AutoAnswerCallPay, "{1}", "Rhystic Study", "Bob paid {1} for Rhystic Study (automatic)"},
		{game.AutoAnswerCallDontPay, "", "Rhystic Study", "Bob didn't pay for Rhystic Study (automatic)"},
		{game.AutoAnswerCallYes, "Consecrated Sphinx — draw two cards", "Consecrated Sphinx",
			"Bob answered Yes to Consecrated Sphinx — draw two cards (automatic)"},
		{game.AutoAnswerCallNo, "Consecrated Sphinx — draw two cards", "Consecrated Sphinx",
			"Bob answered No to Consecrated Sphinx — draw two cards (automatic)"},
		{game.AutoAnswerCallYes, "", "Rhystic Study", "Bob answered Yes to Rhystic Study (automatic)"},
		// Redacted: the name and the label went together.
		{game.AutoAnswerCallPay, "", "", "Bob paid for a card (automatic)"},
		{game.AutoAnswerCallYes, "", "", "Bob answered Yes to a card (automatic)"},
	} {
		e := LogEvent{Kind: LogAutoAnswer, Call: c.call, Label: c.label, actorName: "Bob"}
		if got := renderLogText(e, c.card, ""); got != c.want {
			t.Errorf("%s/%q: got %q, want %q", c.call, c.label, got, c.want)
		}
	}
}

func TestAutoAnswerEventProjectsToTheLog(t *testing.T) {
	g := newTwoSeatGame(t)
	a := g.Seats[0]
	study := game.Card{InstanceID: uuid.New(), Name: "Rhystic Study", TypeLine: "Enchantment", Owner: a.ID, Controller: a.ID,
		KnownBy: map[uuid.UUID]bool{a.ID: true, g.Seats[1].ID: true}}
	g.Battlefield.PushTop(study)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAutoAnswer, Actor: a.ID, Source: study.InstanceID,
			Call: game.AutoAnswerCallPay, Label: "{1}"})
	})
	view := FilterViewFor(ViewOfGame(g), g.Seats[1].ID.String())
	var found *LogEvent
	for i := range view.Log {
		if view.Log[i].Kind == LogAutoAnswer {
			found = &view.Log[i]
		}
	}
	if found == nil {
		t.Fatal("no auto_answer line in the log")
	}
	want := a.Name + " paid {1} for Rhystic Study (automatic)"
	if found.Text != want || found.Call != game.AutoAnswerCallPay {
		t.Errorf("line = %q (call %q), want %q", found.Text, found.Call, want)
	}
}

func TestAutoAnswerViewFieldsArePrivate(t *testing.T) {
	g := newTwoSeatGame(t)
	a, b := g.Seats[0], g.Seats[1]
	if err := g.SetAutoAnswers(a.ID, map[string]game.AutoAnswer{"k2": game.AutoAnswerNever, "k1": game.AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoicePayUnless, Chooser: a.ID, Count: 1, PayCost: "{1}",
			AutoAnswerKey: "k1", AutoAnswerCard: "Rhystic Study", AutoAnswerPrompt: "Rhystic Study — pay {1}?",
			AskedByHand: game.AskedByHandNoMana,
		})
	})
	raw := ViewOfGame(g)
	mine := FilterViewFor(raw, a.ID.String())
	if got := mine.Seats[0].AutoAnswers; len(got) != 2 || got[0].Key != "k1" || got[0].Answer != "always" {
		t.Errorf("own rules = %+v, want two, sorted by key", got)
	}
	pc := mine.PendingChoices[len(mine.PendingChoices)-1]
	if pc.AutoAnswerKey != "k1" || pc.AutoAnswerCard != "Rhystic Study" ||
		pc.AutoAnswerPrompt != "Rhystic Study — pay {1}?" || pc.AskedByHand != "no_mana" {
		t.Errorf("chooser's prompt = %+v", pc)
	}
	for _, viewer := range []string{b.ID.String(), SpectatorViewerID} {
		v := FilterViewFor(raw, viewer)
		if len(v.Seats[0].AutoAnswers) != 0 {
			t.Errorf("viewer %s sees seat A's rules", viewer)
		}
		pc := v.PendingChoices[len(v.PendingChoices)-1]
		if pc.AutoAnswerKey != "" || pc.AutoAnswerCard != "" || pc.AutoAnswerPrompt != "" || pc.AskedByHand != "" {
			t.Errorf("viewer %s sees the chooser's fields: %+v", viewer, pc)
		}
	}
}
