package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_order_memory_test.go — ADR 0018's eighth amendment (#3016):
// the log line every seat reads, and the fields only the chooser gets.

// A remembered trigger order's line names the rows in the order they
// resolve.
func TestTriggerOrderAutoAnswerLogLine(t *testing.T) {
	e := LogEvent{Kind: LogAutoAnswer, Call: game.AutoAnswerCallOrder,
		Label: "Pitiless Plunderer — create a Treasure, then Falkenrath Noble — drain 1", actorName: "Luke"}
	want := "Luke ordered Pitiless Plunderer — create a Treasure, then Falkenrath Noble — drain 1 (automatic)"
	if got := renderLogText(e, "", ""); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// The remembered flag, the rows and the memory count are the chooser's
// alone.
func TestTriggerOrderMemoryFieldsAreTheChoosers(t *testing.T) {
	g := newTwoSeatGame(t)
	me, other := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		for _, row := range []string{"plunderer", "noble", "plunderer"} {
			g.PendingTriggers = append(g.PendingTriggers, &game.StackItem{
				ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID,
				SourceCardID: uuid.New(), Label: row, Body: "catalog/triggered",
				Effect: func(*game.Game, *game.StackItem) error { return nil },
				Params: game.EffectParams{Ability: &game.AbilityRef{Key: row, Slot: game.AbilitySlotTriggered, Ref: "own:0", Name: row}},
			})
		}
	})
	g.RunStateChecksForTest()
	var key string
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceTriggerOrder {
			key = c.AutoAnswerKey
		}
	}
	if key == "" {
		t.Fatal("no keyed trigger-order prompt")
	}
	me.TriggerOrderMemory = map[string]game.RememberedTriggerOrder{key: {Rows: []string{"x"}, Used: 1}}
	view := ViewOfGame(g)
	pick := func(v GameView) *PendingChoiceView {
		for i := range v.PendingChoices {
			if v.PendingChoices[i].Kind == string(game.PendingChoiceTriggerOrder) {
				return &v.PendingChoices[i]
			}
		}
		t.Fatal("no trigger_order prompt in the view")
		return nil
	}
	mine := FilterViewFor(view, me.ID.String())
	theirs := FilterViewFor(view, other.ID.String())
	mp, tp := pick(mine), pick(theirs)
	if !mp.TriggerOrderRemembered || mp.AutoAnswerKey != key {
		t.Errorf("chooser's prompt: remembered %v key %q", mp.TriggerOrderRemembered, mp.AutoAnswerKey)
	}
	var rows []int
	for _, o := range mp.TriggerOptions {
		rows = append(rows, o.Row)
	}
	if len(rows) != 3 || rows[0] != 1 || rows[1] != 2 || rows[2] != 1 {
		t.Errorf("chooser's rows = %v, want [1 2 1]", rows)
	}
	if tp.TriggerOrderRemembered || tp.AutoAnswerKey != "" {
		t.Error("another seat sees the remembered order or its key")
	}
	for _, o := range tp.TriggerOptions {
		if o.Row != 0 {
			t.Errorf("another seat sees a row: %+v", o)
		}
	}
	if mine.Seats[me.Seat].TriggerOrderMemoryCount != 1 || theirs.Seats[me.Seat].TriggerOrderMemoryCount != 0 {
		t.Error("the memory count is not private to its seat")
	}
}
