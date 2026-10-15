package ws

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// trigger_order_memory_test.go — ADR 0018's eighth amendment (#3016)
// at the room: a remembered trigger order answers a repeat batch as an
// ADR 0127 automatic answer, the chooser's and free to undo, and an
// undo reopens the prompt to be asked by hand with the memory kept.

// queueRowsForRoom queues one stamped catalog trigger per row for owner
// and runs the drain, raising the trigger-order prompt.
func queueRowsForRoom(g *game.Game, owner uuid.UUID, rows ...string) func() error {
	return func() error {
		g.WithWriteLock(func() {
			for _, row := range rows {
				g.PendingTriggers = append(g.PendingTriggers, &game.StackItem{
					ID:           uuid.New(),
					Kind:         game.StackItemTriggered,
					Controller:   owner,
					Owner:        owner,
					SourceCardID: uuid.New(),
					Label:        row + " trigger",
					Effect:       func(*game.Game, *game.StackItem) error { return nil },
					Body:         "catalog/triggered",
					Params:       game.EffectParams{Ability: &game.AbilityRef{Key: row, Slot: game.AbilitySlotTriggered, Ref: "own:0", Name: row}},
				})
			}
		})
		g.RunStateChecksForTest()
		return nil
	}
}

func triggerOrderPromptOf(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestRoomAppliesARememberedTriggerOrder(t *testing.T) {
	room, g := newAutoAnswerRoom(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	if _, _, err := room.Apply(me.ID, queueRowsForRoom(g, me.ID, "plunderer", "noble")); err != nil {
		t.Fatal(err)
	}
	prompt := triggerOrderPromptOf(g, me.ID)
	if prompt == nil || prompt.AutoAnswerKey == "" {
		t.Fatalf("want a keyed prompt, got %+v", prompt)
	}
	// Noble first, and remember it.
	order := []uuid.UUID{prompt.TriggerOrderIDs[1], prompt.TriggerOrderIDs[0]}
	remember := true
	if _, _, err := room.Apply(me.ID, func() error {
		return g.ResolveTriggerOrderRemember(prompt.ID, me.ID, order, &remember)
	}); err != nil {
		t.Fatal(err)
	}

	// The same rows again, raised by the other seat's commit.
	view, _, err := room.Apply(other.ID, queueRowsForRoom(g, me.ID, "noble", "plunderer", "noble"))
	if err != nil {
		t.Fatal(err)
	}
	if triggerOrderPromptOf(g, me.ID) != nil {
		t.Fatal("the repeat batch was not answered")
	}
	top := room.undoStack[len(room.undoStack)-1]
	if top.caller != me.ID || !top.freeUndo || top.autoAnswered == uuid.Nil {
		t.Fatalf("top entry = caller %v free %v answered %v; want the chooser's, free", top.caller, top.freeUndo, top.autoAnswered)
	}
	var line *protocol.LogEvent
	for i := range view.Log {
		if view.Log[i].Kind == protocol.LogAutoAnswer {
			line = &view.Log[i]
		}
	}
	if line == nil || line.Call != game.AutoAnswerCallOrder {
		t.Fatalf("no trigger-order auto_answer line: %+v", line)
	}
	if !strings.Contains(line.Text, "ordered noble trigger, then plunderer trigger (automatic)") {
		t.Errorf("log text = %q", line.Text)
	}
	mine := protocol.FilterViewFor(view, me.ID.String())
	theirs := protocol.FilterViewFor(view, other.ID.String())
	if mine.Seats[me.Seat].TriggerOrderMemoryCount != 1 || theirs.Seats[me.Seat].TriggerOrderMemoryCount != 0 {
		t.Errorf("memory count: own %d, other's %d; want 1 and private",
			mine.Seats[me.Seat].TriggerOrderMemoryCount, theirs.Seats[me.Seat].TriggerOrderMemoryCount)
	}

	// Undo: the prompt is back, asked by hand, and the memory stays.
	if view, _, err = room.Undo(me.ID); err != nil {
		t.Fatal(err)
	}
	reopened := triggerOrderPromptOf(g, me.ID)
	if reopened == nil || reopened.AskedByHand != game.AskedByHandUndone {
		t.Fatalf("after undo: prompt %+v, want it asked by hand", reopened)
	}
	mine = protocol.FilterViewFor(view, me.ID.String())
	var pv *protocol.PendingChoiceView
	for i := range mine.PendingChoices {
		if mine.PendingChoices[i].Kind == string(game.PendingChoiceTriggerOrder) {
			pv = &mine.PendingChoices[i]
		}
	}
	if pv == nil || !pv.TriggerOrderRemembered || pv.AskedByHand != string(game.AskedByHandUndone) {
		t.Fatalf("chooser's view of the reopened prompt = %+v", pv)
	}
}
