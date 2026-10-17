package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// venture_test.go — ADR 0203: the bot choosing the next Undercity room.

func ventureInput(t *testing.T, board []protocol.CardView, rooms ...string) aiseat.Input {
	t.Helper()
	const choiceID = "choice-venture"
	labels := map[string]string{
		"forge":     "Forge — Put two +1/+1 counters on target creature.",
		"lost-well": "Lost Well — Scry 2.",
		"trap":      "Trap! — Target player loses 5 life.",
		"arena":     "Arena — Goad target creature.",
		"stash":     "Stash — Create a Treasure token.",
		"archives":  "Archives — Draw a card.",
		"catacombs": "Catacombs — Create a 4/1 black Skeleton creature token with menace.",
	}
	var options []protocol.PickOptionView
	for _, r := range rooms {
		options = append(options, protocol.PickOptionView{Label: labels[r], Dungeon: "undercity", Room: r})
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(board...),
		withChoice(protocol.PendingChoiceView{
			ID:          choiceID,
			Kind:        "option_pick",
			Chooser:     seatID(0).String(),
			Reason:      "venture: choose the next room",
			PickOptions: options,
		}))
	var moves []legal.Move
	for i, o := range options {
		moves = append(moves, choiceMove(t, 0, choiceID, o.Label, map[string]any{"option_index": i}))
	}
	return input(0, v, moves...)
}

// With a creature of its own, Forge's counters are worth the most out
// of Secret Entrance.
func TestTheBotForgesWithACreatureToTakeTheCounters(t *testing.T) {
	in := ventureInput(t, []protocol.CardView{creature("bear", 0, "Bear", 2, 2)}, "lost-well", "forge")
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Forge — Put two +1/+1 counters on target creature." {
		t.Errorf("the bot chose %q, want Forge", got)
	}
}

// With no creature of its own, Forge would put the counters on an
// opponent's creature, so the bot takes Lost Well.
func TestTheBotSkipsForgeWithNoCreatureOfItsOwn(t *testing.T) {
	in := ventureInput(t, []protocol.CardView{creature("theirs", 1, "Their Bear", 2, 2)}, "forge", "lost-well")
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Lost Well — Scry 2." {
		t.Errorf("the bot chose %q, want Lost Well", got)
	}
}

// Trap! beats Arena out of Forge.
func TestTheBotTakesTrapOverArena(t *testing.T) {
	in := ventureInput(t, []protocol.CardView{creature("theirs", 1, "Their Bear", 2, 2)}, "arena", "trap")
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Trap! — Target player loses 5 life." {
		t.Errorf("the bot chose %q, want Trap!", got)
	}
}

// Arena with nothing of an opponent's to goad would goad its own
// creature; Stash is the better room out of Lost Well.
func TestTheBotSkipsArenaWithNothingToGoad(t *testing.T) {
	in := ventureInput(t, []protocol.CardView{creature("bear", 0, "Bear", 2, 2)}, "arena", "stash")
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Stash — Create a Treasure token." {
		t.Errorf("the bot chose %q, want Stash", got)
	}
}
