package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestCoinCallsEnumerateAndDispatchEveryAnswer(t *testing.T) {
	for _, stop := range []bool{false, true} {
		g := newTable(t)
		me := g.Seats[0]
		var id uuid.UUID
		g.WithWriteLock(func() { id = g.FlipCoinForEffect(game.CoinFlipSpec{Flipper: me.ID, Coins: 3, AllowStop: stop}) })
		moves := legal.EnumerateFor(g, me.ID)
		want := 2
		if stop {
			want = 3
		}
		if len(moves) != want {
			t.Fatalf("stop=%v: %d moves, want %d", stop, len(moves), want)
		}
		for _, m := range moves {
			var p struct {
				ChoiceID string `json:"choice_id"`
				Call     string `json:"call"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if p.ChoiceID != id.String() || (p.Call != "heads" && p.Call != "tails" && !(stop && p.Call == "stop")) {
				t.Fatalf("bad answer: %+v", p)
			}
			if m.AlwaysLegal != (p.Call == "heads") {
				t.Fatalf("unexpected always-legal flag: %s", p.Call)
			}
		}
		dispatchAll(t, g, me.ID, moves)
	}
}
