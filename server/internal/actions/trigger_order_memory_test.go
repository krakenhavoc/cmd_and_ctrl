package actions

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_order_memory_test.go — ADR 0018's eighth amendment (#3016) at
// the dispatch layer: resolve_choice's `remember` and
// forget_trigger_order.

func TestRememberAndForgetTriggerOrderThroughDispatch(t *testing.T) {
	g := newGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	var ids []string
	g.WithWriteLock(func() {
		for _, row := range []string{"plunderer", "noble"} {
			it := &game.StackItem{
				ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID,
				SourceCardID: uuid.New(), Label: row, Body: "catalog/triggered",
				Effect: func(*game.Game, *game.StackItem) error { return nil },
				Params: game.EffectParams{Ability: &game.AbilityRef{Key: row, Slot: game.AbilitySlotTriggered, Ref: "own:0", Name: row}},
			}
			ids = append(ids, it.ID.String())
			g.PendingTriggers = append(g.PendingTriggers, it)
		}
	})
	g.RunStateChecksForTest()
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c.Kind == game.PendingChoiceTriggerOrder {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no trigger-order prompt")
	}
	params := `{"choice_id":"` + prompt.ID.String() + `","order":["` + ids[1] + `","` + ids[0] + `"],"remember":true}`
	if err := Dispatch(g, Action{Type: TypeResolveChoice, Player: me.ID, Caller: me.ID, Params: []byte(params)}); err != nil {
		t.Fatal(err)
	}
	if len(me.TriggerOrderMemory) != 1 {
		t.Fatalf("remember: true stored %v", me.TriggerOrderMemory)
	}

	forget := func(player, caller uuid.UUID, params string) error {
		return Dispatch(g, Action{Type: TypeForgetTriggerOrder, Player: player, Caller: caller, Params: []byte(params)})
	}
	if err := forget(me.ID, other.ID, `{}`); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("another seat forgot my orders: %v", err)
	}
	if err := forget(me.ID, me.ID, `{"key":"`+strings.Repeat("k", game.MaxForgetTriggerOrderKeyLen+1)+`"}`); err == nil {
		t.Fatal("accepted an oversized key")
	}
	if err := forget(me.ID, me.ID, `{"key":"trigger_order|nothing"}`); err != nil || len(me.TriggerOrderMemory) != 1 {
		t.Fatalf("forgetting an unknown key: err %v, memory %v", err, me.TriggerOrderMemory)
	}
	key, _ := json.Marshal(prompt.AutoAnswerKey)
	if err := forget(me.ID, me.ID, `{"key":`+string(key)+`}`); err != nil || len(me.TriggerOrderMemory) != 0 {
		t.Fatalf("forget by key: err %v, memory %v", err, me.TriggerOrderMemory)
	}
	if err := forget(me.ID, me.ID, ``); err != nil {
		t.Fatalf("forget all with no params: %v", err)
	}
	if !MintsNoUndo(g, TypeForgetTriggerOrder) {
		t.Error("forget_trigger_order mints an undo entry")
	}
}
