package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// resolution_pause_test.go — the bot half of #1289.
//
// A resolution paused on one of its own prompts now holds state-based
// actions and the trigger drain (CR 704.3). That is only safe if the
// pause is always answerable: the seat that owes the prompt must be
// offered its answers, every other seat must be offered nothing (the
// table is blocked, as it always was), and the answer a bot actually
// sends must finish the resolution and hand the table back.

// pausingCounterProbe rewrites any +1/+1 placement. Two of them with
// different rewrites make the CR 614 counter window queue the CR 616
// ordering prompt: a Doubling Season beside a Hardened Scales.
func pausingCounterProbe(label string, rewrite func(int) int) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventCounter && ev.CounterName == game.CounterPlusOne && ev.CounterDelta > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.CounterDelta = rewrite(ev.CounterDelta)
			return nil
		},
		Label: label,
	}
}

func TestAPausedResolutionIsAnsweredByTheBotAndHandsTheTableBack(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	land := battlefieldCard(g, me, basic("Forest", "Forest"))

	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(pausingCounterProbe("twice", func(n int) int { return n * 2 }))
		g.RegisterReplacementForTest(pausingCounterProbe("one more", func(n int) int { return n + 1 }))
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		id := uuid.New()
		g.StackMeta[id] = &game.StackItem{
			ID: id, Kind: game.StackItemTriggered, Controller: me.ID, Owner: me.ID,
			Label: "earthbend 3",
			Effect: func(g *game.Game, it *game.StackItem) error {
				return g.EarthbendForEffect(it.Controller, uuid.Nil, land, 3)
			},
		}
	})
	for i := 0; i < len(g.Seats) && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.ResolutionPaused() {
		t.Fatal("the earthbend did not pause on its CR 616 prompt")
	}

	for _, p := range g.Seats {
		moves := legal.EnumerateFor(g, p.ID)
		if p.ID != me.ID {
			if len(moves) != 0 {
				t.Errorf("%s was offered %v while another seat's resolution is paused", p.Name, labels(moves))
			}
			continue
		}
		if len(moves) == 0 {
			t.Fatal("the seat that owes the prompt was offered nothing: the table is wedged")
		}
		for _, m := range moves {
			if m.Kind != legal.KindChoice {
				t.Errorf("owing seat offered %q (%s), want only the prompt's answers", m.Label, m.Kind)
			}
		}
		dispatchAll(t, g, me.ID, moves)
	}

	// The bot's move, sent the way the runner sends it.
	m := legal.EnumerateFor(g, me.ID)[0]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params,
	}); err != nil {
		t.Fatalf("dispatch %q: %v", m.Label, err)
	}

	if g.ResolutionPaused() {
		t.Fatal("the resolution is still paused after the bot answered")
	}
	var landCard *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == land {
			landCard = &g.Battlefield.Cards[i]
		}
	}
	if landCard == nil || landCard.Counters[game.CounterPlusOne] == 0 {
		t.Fatal("the earthbent land is gone or has no counters: the sweep ran inside the resolution")
	}
	if !hasKind(legal.EnumerateFor(g, g.Seats[g.Turn.PriorityHolder].ID), legal.KindPass) {
		t.Error("the priority holder has no pass after the resolution finished")
	}
}

func hasKind(moves []legal.Move, k legal.Kind) bool {
	for _, m := range moves {
		if m.Kind == k {
			return true
		}
	}
	return false
}
