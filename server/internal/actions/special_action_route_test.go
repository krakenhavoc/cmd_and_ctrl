package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ADR 0189 decision 4: special_action.route is parsed at the door, and
// an unknown value is refused before the engine is asked anything.
func TestSpecialActionRouteIsParsed(t *testing.T) {
	g := newGame(t)
	me := g.Seats[g.Turn.PriorityHolder]
	for _, route := range []string{"megamorph", "Morph", "mana-cost"} {
		err := Dispatch(g, Action{
			Type: TypeSpecialAction, Player: me.ID, Caller: me.ID,
			Params: params(t, map[string]any{"card_id": uuid.New().String(), "kind": "turn_face_up", "route": route}),
		})
		if err == nil || !strings.Contains(err.Error(), "route") {
			t.Errorf("route %q: err = %v, want a route refusal from the parser", route, err)
		}
	}
	// A known route passes the parser and reaches the engine, which
	// does not find the card.
	err := Dispatch(g, Action{
		Type: TypeSpecialAction, Player: me.ID, Caller: me.ID,
		Params: params(t, map[string]any{"card_id": uuid.New().String(), "kind": "turn_face_up", "route": "morph"}),
	})
	if err == nil || strings.Contains(err.Error(), "special_action route") {
		t.Errorf("a valid route was refused by the parser: %v", err)
	}
}
