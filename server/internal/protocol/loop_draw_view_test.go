package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loop_draw_view_test.go — CR 732.4 on the wire (#2506): a mandatory
// loop that repeats unchanged marks its notice mandatory, then ends the
// game in a draw whose outcome and log line name the loop's card.

const (
	viewOracleEbb  = "test-2506-view-ebb"
	viewOracleFlow = "test-2506-view-flow"
)

// viewLifeTrigger: "whenever you gain (sign > 0) or lose (sign < 0)
// life, change your life by delta".
func viewLifeTrigger(label string, sign, delta int) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.Target == source.Controller && ev.Amount*sign > 0
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			id, controller := source.InstanceID, source.Controller
			item := game.NewTriggeredItem(source, label)
			item.Effect = func(g *game.Game, _ *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(id, controller, delta)
			}
			return item
		},
	}
}

func TestAMandatoryLoopDrawIsOnTheWire(t *testing.T) {
	prev := game.CatalogTriggers
	game.CatalogTriggers = func(key string) []game.TriggeredAbility {
		switch key {
		case viewOracleEbb:
			return []game.TriggeredAbility{viewLifeTrigger("Ebb Engine — lose 1 life", 1, -1)}
		case viewOracleFlow:
			return []game.TriggeredAbility{viewLifeTrigger("Flow Engine — gain 1 life", -1, 1)}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogTriggers = prev })

	g := newTwoSeatGame(t)
	g.LoopThreshold = 3
	me := g.Seats[0]
	ebb, flow := game.NewCard("Ebb Engine", me.ID), game.NewCard("Flow Engine", me.ID)
	ebb.OracleID, flow.OracleID = viewOracleEbb, viewOracleFlow
	ebb.TypeLine, flow.TypeLine = "Enchantment", "Enchantment"
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(ebb)
		g.Battlefield.PushTop(flow)
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})

	sawMandatory := false
	for i := 0; i < 200 && g.State == game.StateActive; i++ {
		if v := ViewOfGame(g); v.LoopNotice != nil {
			if !v.LoopNotice.Mandatory {
				t.Fatalf("loop_notice %+v is not marked mandatory", v.LoopNotice)
			}
			sawMandatory = true
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !sawMandatory {
		t.Error("no seat was ever shown the mandatory notice before the draw")
	}
	v := ViewOfGame(g)
	if v.State != "ended" || v.Outcome == nil || v.Outcome.Kind != "draw" || v.Outcome.Cause != "mandatory_loop" ||
		v.Outcome.Winner != "" || v.Outcome.SourceName != "Flow Engine" {
		t.Fatalf("state %s, outcome %+v", v.State, v.Outcome)
	}
	if v.LoopNotice != nil {
		t.Errorf("loop_notice %+v on an ended game", v.LoopNotice)
	}
	over := logOfKind(v, LogGameOver)
	want := "The game is a draw: Flow Engine's loop repeats with no way to stop (CR 732.4)"
	if len(over) != 1 || over[0].Text != want || over[0].Seat != NoSeat || over[0].Cause != "mandatory_loop" {
		t.Errorf("game_over lines %+v, want one %q", over, want)
	}
}
