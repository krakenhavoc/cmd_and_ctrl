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

// ventureOptionsInput is a venture prompt with these options, for a
// seat 0 built with `me`.
func ventureOptionsInput(t *testing.T, me []seatOpt, board []protocol.CardView, reason string, options ...protocol.PickOptionView) aiseat.Input {
	t.Helper()
	const choiceID = "choice-venture"
	v := newView([]protocol.PlayerView{newSeat(0, me...), newSeat(1)},
		withBattlefield(board...),
		withChoice(protocol.PendingChoiceView{
			ID:          choiceID,
			Kind:        "option_pick",
			Chooser:     seatID(0).String(),
			Reason:      reason,
			PickOptions: options,
		}))
	var moves []legal.Move
	for i, o := range options {
		moves = append(moves, choiceMove(t, 0, choiceID, o.Label, map[string]any{"option_index": i}))
	}
	return input(0, v, moves...)
}

// dungeonOptions are the three dungeons a plain venture offers (#4100).
var dungeonOptions = []protocol.PickOptionView{
	{Label: "Lost Mine of Phandelver", Dungeon: "lost-mine-of-phandelver"},
	{Label: "Dungeon of the Mad Mage", Dungeon: "dungeon-of-the-mad-mage"},
	{Label: "Tomb of Annihilation", Dungeon: "tomb-of-annihilation"},
}

// With life to spare the bot takes the fastest dungeon, the Tomb; low on
// life it takes Lost Mine of Phandelver, whose rooms cost it nothing.
func TestTheBotChoosesADungeonByItsLife(t *testing.T) {
	in := ventureOptionsInput(t, nil, nil, "venture: choose a dungeon", dungeonOptions...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Tomb of Annihilation" {
		t.Errorf("at 40 life the bot chose %q, want Tomb of Annihilation", got)
	}
	in = ventureOptionsInput(t, []seatOpt{withLife(8)}, nil, "venture: choose a dungeon", dungeonOptions...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Lost Mine of Phandelver" {
		t.Errorf("at 8 life the bot chose %q, want Lost Mine of Phandelver", got)
	}
}

// Out of Mine Tunnels: Fungi Cavern only with an enemy creature to
// shrink, otherwise Dark Pool.
func TestTheBotShrinksOnlyAnEnemyInFungiCavern(t *testing.T) {
	opts := []protocol.PickOptionView{
		{Label: "Dark Pool — Each opponent loses 1 life and you gain 1 life.", Dungeon: "lost-mine-of-phandelver", Room: "dark-pool"},
		{Label: "Fungi Cavern — Target creature gets -4/-0 until your next turn.", Dungeon: "lost-mine-of-phandelver", Room: "fungi-cavern"},
	}
	in := ventureOptionsInput(t, nil, []protocol.CardView{creature("mine", 0, "Bear", 2, 2)}, "venture: choose the next room", opts...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != opts[0].Label {
		t.Errorf("with only its own creature the bot chose %q, want Dark Pool", got)
	}
}

// Out of Trapped Entry: Oubliette when the bot has nothing to lose to
// it, Veils of Fear when it would give up a creature and a land.
func TestTheBotTakesOublietteOnlyWhenItIsCheap(t *testing.T) {
	opts := []protocol.PickOptionView{
		{Label: "Veils of Fear — Each player loses 2 life unless they discard a card.", Dungeon: "tomb-of-annihilation", Room: "veils-of-fear"},
		{Label: "Oubliette — Discard a card and sacrifice a creature, an artifact, and a land.", Dungeon: "tomb-of-annihilation", Room: "oubliette"},
	}
	in := ventureOptionsInput(t, nil, nil, "venture: choose the next room", opts...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != opts[1].Label {
		t.Errorf("with an empty board and hand the bot chose %q, want Oubliette", got)
	}
	board := []protocol.CardView{creature("mine", 0, "Bear", 2, 2), land("mountain", 0)}
	in = ventureOptionsInput(t, nil, board, "venture: choose the next room", opts...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != opts[0].Label {
		t.Errorf("with a creature and a land the bot chose %q, want Veils of Fear", got)
	}
}
