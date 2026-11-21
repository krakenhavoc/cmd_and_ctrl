package effectstest

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scheduledControls is the controls still waiting for their turn.
func ScheduledControls(g *game.Game) []game.PlayerControl {
	var out []game.PlayerControl
	g.ReadSnapshot(func() { out = g.ScheduledPlayerControlsLocked() })
	return out
}

// playLandAsController has `controller` play a land from `controlled`'s
// hand as them, through the dispatcher, which only a live control
// allows (CR 723.5). It returns the land's ID.
func playLandAsController(t *testing.T, g *game.Game, controller, controlled *game.Player) uuid.UUID {
	t.Helper()
	ToMain(t, g)
	land := game.NewCard("Forest", controlled.ID)
	land.TypeLine = "Basic Land — Forest"
	g.WithWriteLock(func() { controlled.Hand.PushTop(land) })
	raw, err := json.Marshal(map[string]any{"instance_id": land.InstanceID.String()})
	if err != nil {
		t.Fatal(err)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeCastSpell, Player: controlled.ID, Caller: controller.ID, Params: raw}); err != nil {
		t.Fatalf("%s playing %s's land: %v", controller.Name, controlled.Name, err)
	}
	if c, ok := g.LookupCardForEffect(land.InstanceID); !ok || !g.Battlefield.Contains(land.InstanceID) || c.Controller != controlled.ID {
		t.Fatalf("the land is not on the battlefield under %s's control: %+v", controlled.Name, c)
	}
	return land.InstanceID
}

// walkControlledTurnWithNoExtraTurn ends turns until `controlled`'s turn
// begins, checks the controller holds it and can play as them, ends it,
// and checks the control is gone and the table moved on to the seat
// after `controlled` with no extra turn between (the difference from
// Emrakul, the Promised End).
func WalkControlledTurnWithNoExtraTurn(t *testing.T, g *game.Game, controller, controlled *game.Player) {
	t.Helper()
	n := len(g.Seats)
	for i := 0; g.Turn.ActiveSeat != controlled.Seat; i++ {
		if i > n {
			t.Fatalf("never reached %s's turn", controlled.Name)
		}
		if got := g.ControllerOfPlayer(controlled.ID); got != controlled.ID {
			t.Fatalf("the control applies before %s's turn: %v", controlled.Name, got)
		}
		EndTurn(t, g)
	}
	if g.Turn.Extra {
		t.Fatal("the controlled turn is an extra turn")
	}
	if got := g.ControllerOfPlayer(controlled.ID); got != controller.ID {
		t.Fatalf("%s's turn is controlled by %v, want %s", controlled.Name, got, controller.Name)
	}
	if got := g.ControlUntilOfPlayer(controlled.ID); got != game.ControlUntilTurn {
		t.Errorf("control lasts %q, want a whole turn", got)
	}
	playLandAsController(t, g, controller, controlled)
	// The controlled player decides nothing of their own turn (CR 723.5).
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypePassPriority, Player: controlled.ID, Caller: controlled.ID}); err == nil {
		t.Error("the controlled player passed priority for themselves")
	}

	EndTurn(t, g)
	AssertTurnOf(t, g, (controlled.Seat+1)%n, false, "after the controlled turn: the rotation resumes, no extra turn")
	if got := g.ControllerOfPlayer(controlled.ID); got != controlled.ID {
		t.Errorf("the control outlived the turn: %v", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left after the controlled turn: %+v", g.PlayerControls)
	}
}
