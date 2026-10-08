package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// auto_answer_test.go — ADR 0127 §4: the server answers a prompt only
// when the enumerator would list an answer to it for its chooser now,
// so an automatic answer goes in the order the table would take it.
// game.NextAutoAnswer states that condition itself (game cannot import
// legal); this pins the two to agreeing. And §3: set_auto_answers is
// never a move.
func TestAutoAnswerOnlyAnswersWhatTheEnumeratorOffers(t *testing.T) {
	g := newTable(t)
	payer := g.Seats[1]
	g.WithWriteLock(func() {
		_ = g.QueuePayUnlessForEffect(payer.ID, uuid.Nil, "{1}", "Test — pay {1}?", nil)
		g.PendingChoices[len(g.PendingChoices)-1].AutoAnswerKey = "tax"
	})
	if err := g.SetAutoAnswers(payer.ID, map[string]game.AutoAnswer{"tax": game.AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	id, chooser, ok := g.NextAutoAnswer()
	if !ok || chooser != payer.ID {
		t.Fatalf("NextAutoAnswer = %v %v %v", id, chooser, ok)
	}
	offered := false
	for _, m := range legal.EnumerateFor(g, chooser) {
		if strings.Contains(string(m.Params), id.String()) {
			offered = true
		}
		if m.Type == "set_auto_answers" {
			t.Errorf("set_auto_answers offered as a move: %+v", m)
		}
	}
	if !offered {
		t.Error("the prompt NextAutoAnswer named has no answer in its chooser's move list")
	}
}
