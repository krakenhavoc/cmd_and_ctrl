package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// trigger_order_memory_test.go — ADR 0018's eighth amendment (I1): an
// agent seat is offered one extra move on a keyed trigger-order prompt,
// the order shown with each row's copies together and `remember: true`.
// A bot and a person are not; forget_trigger_order is never a move.
func TestAgentSeatIsOfferedTheRememberMove(t *testing.T) {
	for _, c := range []struct {
		name       string
		agent, bot bool
		want       int
	}{
		{"agent", true, false, 1},
		{"person", false, false, 0},
		{"bot", false, true, 0},
	} {
		g := newTable(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
			t.Fatal(err)
		}
		me.Agent, me.IsBot = c.agent, c.bot
		var ids []uuid.UUID
		g.WithWriteLock(func() {
			for _, row := range []string{"plunderer", "noble", "plunderer"} {
				it := &game.StackItem{
					ID: uuid.New(), Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID,
					SourceCardID: uuid.New(), Label: row, Body: "catalog/triggered",
					Effect: func(*game.Game, *game.StackItem) error { return nil },
					Params: game.EffectParams{Ability: &game.AbilityRef{Key: row, Slot: game.AbilitySlotTriggered, Ref: "own:0", Name: row}},
				}
				ids = append(ids, it.ID)
				g.PendingTriggers = append(g.PendingTriggers, it)
			}
		})
		g.RunStateChecksForTest()
		var remember []legal.Move
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if strings.Contains(string(m.Params), `"remember":true`) {
				remember = append(remember, m)
			}
			if m.Type == "forget_trigger_order" {
				t.Errorf("%s: forget_trigger_order offered as a move", c.name)
			}
		}
		if len(remember) != c.want {
			t.Fatalf("%s: %d remember moves, want %d", c.name, len(remember), c.want)
		}
		if c.want == 0 {
			continue
		}
		m := remember[0]
		want := `"order":["` + ids[0].String() + `","` + ids[2].String() + `","` + ids[1].String() + `"]`
		if !strings.Contains(string(m.Params), want) {
			t.Errorf("params %s, want the order shown with the Plunderers together", m.Params)
		}
		if !strings.HasSuffix(m.Label, "and use this order again") {
			t.Errorf("label %q", m.Label)
		}
	}
}
