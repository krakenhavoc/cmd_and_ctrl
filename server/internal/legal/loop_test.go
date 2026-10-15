package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// loop_test.go — ADR 0193 §8 (CR 732.2): the demonstrated-loop moves.
// An agent seat may record and propose; a bot or a browser seat is not
// offered the recording verbs; and every seat can answer a proposal,
// a bot only ever by accepting.

func loopMovesOf(moves []legal.Move) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindLoop {
			out = append(out, m)
		}
	}
	return out
}

// markSeat sets a seat's kind on a started game (SetBot and SetAgent
// are lobby-only).
func markSeat(g *game.Game, seat uuid.UUID, set func(*game.Player)) {
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.ID == seat {
				set(p)
			}
		}
	})
}

func dispatchAs(t *testing.T, g *game.Game, seat uuid.UUID, typ string, params any) {
	t.Helper()
	var raw json.RawMessage
	if params != nil {
		raw, _ = json.Marshal(params)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(typ), Player: seat, Caller: seat, Params: raw}); err != nil {
		t.Fatalf("%s: %v", typ, err)
	}
}

func TestOnlyAnAgentSeatIsOfferedTheLoopRecordingVerbs(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat].ID
	if got := loopMovesOf(legal.EnumerateFor(g, me)); len(got) != 0 {
		t.Fatalf("a browser seat was offered %v", labels(got))
	}
	markSeat(g, me, func(p *game.Player) { p.IsBot = true })
	if got := loopMovesOf(legal.EnumerateFor(g, me)); len(got) != 0 {
		t.Fatalf("a bot seat was offered %v", labels(got))
	}

	g = newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me = g.Seats[g.Turn.ActiveSeat].ID
	markSeat(g, me, func(p *game.Player) { p.Agent = true })
	got := loopMovesOf(legal.EnumerateFor(g, me))
	if len(got) != 1 || got[0].Type != legal.TypeLoopRecordStart {
		t.Fatalf("an agent seat with priority was offered %v, want record", labels(got))
	}
	dispatchAll(t, g, me, got)

	dispatchAs(t, g, me, legal.TypeLoopRecordStart, nil)
	got = loopMovesOf(legal.EnumerateFor(g, me))
	if len(got) != 1 || got[0].Type != legal.TypeLoopRecordCancel {
		t.Fatalf("with nothing recorded the recorder is offered %v, want cancel", labels(got))
	}
	// One step: a shuffle, which closes.
	dispatchAs(t, g, me, string(actions.TypeShuffleLibrary), nil)
	got = loopMovesOf(legal.EnumerateFor(g, me))
	if len(got) != 2 || got[0].Type != legal.TypeLoopRecordStop {
		t.Fatalf("while recording the recorder is offered %v, want stop and cancel", labels(got))
	}
	dispatchAll(t, g, me, got)

	dispatchAs(t, g, me, legal.TypeLoopRecordStop, nil)
	got = loopMovesOf(legal.EnumerateFor(g, me))
	if len(got) != len(legal.LoopProposeCounts)+1 {
		t.Fatalf("a ready recording offers %v, want a proposal per count and cancel", labels(got))
	}
	dispatchAll(t, g, me, got)
}

func TestABotSeatOnlyAcceptsALoopProposal(t *testing.T) {
	g := newTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	active := g.Turn.ActiveSeat
	me := g.Seats[active].ID
	next := g.Seats[(active+1)%len(g.Seats)].ID
	// Two steps that leave the step where it is, one of them the next
	// seat's own choice: a shuffle, and a life change.
	dispatchAs(t, g, me, legal.TypeLoopRecordStart, nil)
	dispatchAs(t, g, me, string(actions.TypeShuffleLibrary), nil)
	dispatchAs(t, g, next, string(actions.TypeChangeLife), map[string]int{"delta": 1})
	dispatchAs(t, g, me, legal.TypeLoopRecordStop, nil)
	dispatchAs(t, g, me, legal.TypeLoopPropose, map[string]int{"count": 10})

	human := legal.EnumerateFor(g, next)
	if len(human) != 2 || human[0].Type != legal.TypeResolveLoopProposal || !human[0].AlwaysLegal {
		t.Fatalf("the asked seat is offered %v, want an always-legal accept and a shorten", labels(human))
	}
	dispatchAll(t, g, next, human)

	markSeat(g, next, func(p *game.Player) { p.IsBot = true })
	bot := legal.EnumerateFor(g, next)
	if len(bot) != 1 || bot[0].Type != legal.TypeResolveLoopProposal || !bot[0].AlwaysLegal {
		t.Fatalf("a bot seat was offered %v, want accept alone", labels(bot))
	}
	var p struct {
		Accept bool `json:"accept"`
	}
	if err := json.Unmarshal(bot[0].Params, &p); err != nil || !p.Accept {
		t.Fatalf("the bot's move does not accept: %s", bot[0].Params)
	}
	dispatchAll(t, g, next, bot)
}
