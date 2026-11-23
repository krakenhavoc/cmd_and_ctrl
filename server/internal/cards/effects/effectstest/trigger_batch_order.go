package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerTriggerOrderLastQueuedFirst answers an open CR 603.3b prompt so
// the stack comes out exactly as the drain builds it without asking:
// queue order is placement order, so the last-queued item resolves
// first. Before #1529 a targeted trigger always drained last and alone,
// so this is also the stack every test written before the fix observed.
// Reports whether there was a prompt to answer.
func AnswerTriggerOrderLastQueuedFirst(t *testing.T, g *game.Game) bool {
	t.Helper()
	ch := TriggerOrderPrompt(g)
	if ch == nil {
		return false
	}
	ids := ch.TriggerOrderIDs
	order := make([]uuid.UUID, 0, len(ids))
	for i := len(ids) - 1; i >= 0; i-- {
		order = append(order, ids[i])
	}
	if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, order); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
	return true
}

// pendingItemLabelled finds a queued (not yet placed) item by label.
func PendingItemLabelled(g *game.Game, label string, source uuid.UUID) *game.StackItem {
	for _, it := range g.PendingTriggers {
		if it != nil && it.Label == label && it.SourceCardID == source {
			return it
		}
	}
	return nil
}

// answerTriggerOrderResolvingFirst answers the open prompt with the
// item(s) matching `first` resolving before every other item, the rest
// in offered order.
func AnswerTriggerOrderResolvingFirst(t *testing.T, g *game.Game, first uuid.UUID) {
	t.Helper()
	ch := TriggerOrderPrompt(g)
	if ch == nil {
		t.Fatal("no trigger_order prompt to answer")
	}
	order := []uuid.UUID{first}
	for _, id := range ch.TriggerOrderIDs {
		if id != first {
			order = append(order, id)
		}
	}
	if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, order); err != nil {
		t.Fatalf("ResolveTriggerOrder: %v", err)
	}
}
