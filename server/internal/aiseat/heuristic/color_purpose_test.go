package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// color_purpose_test.go is #780: the colour a bot names depends on what
// the card is going to DO with it.
//
// Every test is the same board answered through a different declared
// purpose, because that is the bug: the board never told the old policy
// anything, so a mono-green bot named green for its Coldsteel Heart and
// named green again for its Wash Out, bouncing its own permanents and
// leaving the opposition's alone.
//
// Back the purpose switch out of `colorChoiceValue` and every
// assertion below but the Coldsteel Heart one fails: without it all
// five answers are ranked by the hand's colour need, which is exactly
// one answer for every card.

const colorChoiceID = "00000000-0000-4000-8000-00000000c010"

func colored(c protocol.CardView, colors ...string) protocol.CardView {
	c.Colors = append(append([]string(nil), c.Colors...), colors...)
	return c
}

func colorMoves(t *testing.T, seat int) []legal.Move {
	t.Helper()
	var out []legal.Move
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		out = append(out, choiceMove(t, seat, colorChoiceID, "choose "+c, map[string]any{"color": c}))
	}
	return out
}

func withColorChoice(seat int, purpose string) viewOpt {
	return withChoice(protocol.PendingChoiceView{
		ID:           colorChoiceID,
		Kind:         "choose_color",
		Chooser:      seatID(seat).String(),
		FromPlayer:   seatID(seat).String(),
		Count:        1,
		Reason:       "choose a color",
		ColorOptions: []string{"W", "U", "B", "R", "G"},
		ColorPurpose: purpose,
	})
}

// --- harm: Wash Out ---------------------------------------------------

func TestAMonoGreenBotWashesOutTheOpponentsColour(t *testing.T) {
	// Seat 0 is mono-green: a green board, a green hand. Seat 1 is
	// mono-red with more on the table. Green is the answer the old
	// policy gave for everything.
	green := withHand(colored(spell(cardID(90), 0, "Giant Growth", "{G}"), "G"))
	board := withBattlefield(
		colored(creature(cardID(1), 0, "Llanowar Elf", 1, 1), "G"),
		colored(creature(cardID(2), 0, "Wall of Roots", 0, 5), "G"),
		colored(creature(cardID(3), 1, "Goblin", 2, 2), "R"),
		colored(creature(cardID(4), 1, "Dragon", 5, 5), "R"),
		colored(creature(cardID(5), 1, "Ogre", 4, 4), "R"),
	)
	v := newView([]protocol.PlayerView{newSeat(0, green), newSeat(1)},
		board, withColorChoice(0, "harm"))

	in := input(0, v, colorMoves(t, 0)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose R" {
		t.Errorf("a mono-green bot's Wash Out named %q — it must name the colour that costs the opposition "+
			"most and itself least, not its own", got)
	}
}

func TestHarmDeclinesAColourItWouldLoseMoreOn(t *testing.T) {
	// The mirror: the bot's own green board is the big one, so green is
	// the worst possible name even though the hand is full of it.
	green := withHand(colored(spell(cardID(90), 0, "Giant Growth", "{G}"), "G"))
	v := newView([]protocol.PlayerView{newSeat(0, green), newSeat(1)},
		withBattlefield(
			colored(creature(cardID(1), 0, "Craterhoof", 8, 8), "G"),
			colored(creature(cardID(3), 1, "Goblin", 1, 1), "R"),
		),
		withColorChoice(0, "harm"))

	in := input(0, v, colorMoves(t, 0)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == "choose G" {
		t.Error("the bot named the colour of its own 8/8 for a harm effect")
	}
}

// --- mana: Coldsteel Heart still works -------------------------------

func TestColdsteelHeartStillNamesTheColourTheHandNeeds(t *testing.T) {
	// A blue hand on a green board. The mana pick follows the HAND, or
	// the rock fixes for a colour the bot is not holding.
	hand := withHand(
		colored(spell(cardID(90), 0, "Counterspell", "{U}{U}"), "U"),
		colored(spell(cardID(91), 0, "Brainstorm", "{U}"), "U"),
	)
	board := withBattlefield(
		colored(creature(cardID(1), 0, "Llanowar Elf", 1, 1), "G"),
		colored(creature(cardID(2), 0, "Wall of Roots", 0, 5), "G"),
	)
	for _, purpose := range []string{"mana", ""} {
		t.Run("purpose="+purpose, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0, hand), newSeat(1)},
				board, withColorChoice(0, purpose))
			in := input(0, v, colorMoves(t, 0)...)
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose U" {
				t.Errorf("chose %q, want the colour the hand needs", got)
			}
		})
	}
}

// --- benefit: the board breaks the tie --------------------------------

func TestBenefitFallsBackToTheBotsOwnBoardWhenTheHandIsSilent(t *testing.T) {
	// Selective Obliteration: the colour NAMED is the colour that
	// survives, so with nothing in hand asking for one the bot keeps
	// the colour it has most of on the table.
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			colored(creature(cardID(1), 0, "Bear", 2, 2), "B"),
			colored(creature(cardID(2), 0, "Bear", 2, 2), "B"),
			colored(creature(cardID(3), 0, "Bear", 2, 2), "W"),
			colored(creature(cardID(4), 1, "Theirs", 6, 6), "R"),
		),
		withColorChoice(0, "benefit"))
	in := input(0, v, colorMoves(t, 0)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose B" {
		t.Errorf("chose %q, want the colour the bot has most of — that is the half of its board that survives", got)
	}
}

// --- filter: what the opposition has shown ----------------------------

func TestFilterNamesTheColourTheOppositionHasShownMostOf(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)},
		withBattlefield(
			colored(creature(cardID(1), 0, "Mine", 2, 2), "G"),
			colored(creature(cardID(2), 0, "Mine", 2, 2), "G"),
			colored(creature(cardID(3), 1, "Theirs", 1, 1), "U"),
			colored(creature(cardID(4), 2, "Theirs", 1, 1), "U"),
			colored(creature(cardID(5), 2, "Theirs", 1, 1), "U"),
			colored(creature(cardID(6), 1, "Theirs", 7, 7), "R"),
		),
		withColorChoice(0, "filter"))
	in := input(0, v, colorMoves(t, 0)...)
	// Three blue cards against one red, and the bot's own two green
	// ones do not count — Oona exiles somebody else's library.
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose U" {
		t.Errorf("chose %q, want the colour most common among the opposition's known cards", got)
	}
}

// --- protection: the biggest threat -----------------------------------

func TestMotherOfRunesNamesTheAttackingThreatsColour(t *testing.T) {
	// A red 4/4 already swinging at this seat, and a bigger green 5/5
	// sitting at home. Protection is bought a beat before it is needed,
	// so the one in the red zone is the one to name.
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			colored(creature(cardID(1), 0, "Mine", 1, 1), "W"),
			colored(creature(cardID(2), 1, "Raider", 4, 4, attacking(0)), "R"),
			colored(creature(cardID(3), 1, "Homebody", 5, 5), "G"),
		),
		withTurn(3, 1, "declare_blockers"),
		withColorChoice(0, "protect"))
	in := input(0, v, colorMoves(t, 0)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose R" {
		t.Errorf("chose %q, want the colour of the creature attacking this seat", got)
	}

	// With nothing attacking, the biggest body is the threat.
	quiet := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			colored(creature(cardID(2), 1, "Raider", 4, 4), "R"),
			colored(creature(cardID(3), 1, "Homebody", 5, 5), "G"),
		),
		withColorChoice(0, "protect"))
	in = input(0, quiet, colorMoves(t, 0)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose G" {
		t.Errorf("chose %q on a quiet board, want the colour of the biggest opposing creature", got)
	}
}

// --- the purposes really do differ ------------------------------------

func TestOneBoardAnsweredFiveWays(t *testing.T) {
	// The whole of #780 in one assertion: the same seat, the same
	// board, five prompts, and the answer is not allowed to be the same
	// every time.
	hand := withHand(colored(spell(cardID(90), 0, "Giant Growth", "{G}"), "G"))
	board := withBattlefield(
		colored(creature(cardID(1), 0, "Mine", 3, 3), "G"),
		colored(creature(cardID(2), 1, "Theirs", 4, 4, attacking(0)), "R"),
		colored(creature(cardID(3), 1, "Theirs", 1, 1), "U"),
		colored(creature(cardID(4), 1, "Theirs", 1, 1), "U"),
	)
	answers := map[string]string{}
	for _, purpose := range []string{"mana", "benefit", "harm", "filter", "protect"} {
		v := newView([]protocol.PlayerView{newSeat(0, hand), newSeat(1)},
			board, withTurn(3, 1, "declare_blockers"), withColorChoice(0, purpose))
		in := input(0, v, colorMoves(t, 0)...)
		answers[purpose] = chose(t, in, decide(t, heuristic.New(), in))
	}
	if answers["mana"] != "choose G" || answers["benefit"] != "choose G" {
		t.Errorf("the bot's own colour must still win a mana or benefit prompt: %v", answers)
	}
	if answers["harm"] == "choose G" {
		t.Errorf("a harm prompt named the bot's own colour: %v", answers)
	}
	if answers["filter"] != "choose U" {
		t.Errorf("a filter prompt must name what the opposition has most of: %v", answers)
	}
	if answers["protect"] != "choose R" {
		t.Errorf("a protect prompt must name what is attacking: %v", answers)
	}
}
