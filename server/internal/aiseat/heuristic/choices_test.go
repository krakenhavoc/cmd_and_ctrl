package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func discardMove(t *testing.T, seat int, label string, ids ...string) legal.Move {
	return legal.Move{
		Type: legal.TypeDiscardSelection, Player: seatID(seat), Kind: legal.KindChoice,
		Label: label, Params: mustJSON(t, map[string]any{"card_ids": ids}),
	}
}

// A pending choice is a window the bot MUST answer — the engine
// refuses pass_priority while one is open — so the only question is
// which answer, and the sign of the answer depends on the choice's
// kind. "These card IDs" means pitch the worst for a discard and take
// the best for a search, and a policy that guesses from the payload
// shape alone cannot tell them apart. It reads the kind off
// GameView.PendingChoices instead.
func TestChoicesReadTheKindNotThePayloadShape(t *testing.T) {
	const choiceID = "choice-1"
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	mountain := land(cardID(2), 0)

	t.Run("search takes the best card", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withChoice(protocol.PendingChoiceView{
				ID: choiceID, Kind: "search_library", Chooser: seatID(0).String(),
				Reason: "Search", Options: []protocol.CardView{mountain, dragon},
			}))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "fail to find", map[string]any{"card_ids": []string{}}),
			choiceMove(t, 0, choiceID, "take the land", map[string]any{"card_ids": []string{cardID(2)}}),
			choiceMove(t, 0, choiceID, "take the dragon", map[string]any{"card_ids": []string{cardID(1)}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take the dragon" {
			t.Fatalf("chose %q", got)
		}
	})

	t.Run("sacrifice gives up the least", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(dragon, mountain),
			withChoice(protocol.PendingChoiceView{
				ID: choiceID, Kind: "sacrifice_choice", Chooser: seatID(0).String(), Reason: "Sacrifice",
			}))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "sacrifice the dragon", map[string]any{"card_ids": []string{cardID(1)}}),
			choiceMove(t, 0, choiceID, "sacrifice the land", map[string]any{"card_ids": []string{cardID(2)}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "sacrifice the land" {
			t.Fatalf("chose %q", got)
		}
	})

	t.Run("pick_target points away from us", func(t *testing.T) {
		mine := creature(cardID(10), 0, "Mine", 3, 3)
		theirs := creature(cardID(20), 1, "Theirs", 3, 3)
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(mine, theirs),
			withChoice(protocol.PendingChoiceView{
				ID: choiceID, Kind: "pick_target", Chooser: seatID(0).String(), Reason: "Destroy",
			}))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "destroy mine", map[string]any{"target": cardTarget(cardID(10))}),
			choiceMove(t, 0, choiceID, "destroy theirs", map[string]any{"target": cardTarget(cardID(20))}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "destroy theirs" {
			t.Fatalf("chose %q", got)
		}
	})

	t.Run("scry bottoms what it does not want", func(t *testing.T) {
		// Six lands out: another one is the last thing the bot needs.
		var bf []protocol.CardView
		for i := 0; i < 6; i++ {
			bf = append(bf, land(cardID(30+i), 0))
		}
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(bf...),
			withChoice(protocol.PendingChoiceView{
				ID: choiceID, Kind: "scry", Chooser: seatID(0).String(), Reason: "Scry",
				Options: []protocol.CardView{mountain, dragon},
			}))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "keep all", map[string]any{"top_order": []string{cardID(2), cardID(1)}, "bottom": []string{}}),
			choiceMove(t, 0, choiceID, "bottom the land", map[string]any{"bottom": []string{cardID(2)}, "top_order": []string{cardID(1)}}),
			choiceMove(t, 0, choiceID, "bottom the dragon", map[string]any{"bottom": []string{cardID(1)}, "top_order": []string{cardID(2)}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "bottom the land" {
			t.Fatalf("chose %q", got)
		}
	})

	// #742: Nyx Lotus offers four {G} or one {U}. A hand full of blue
	// symbols must not talk the bot into the smaller pick; with equal
	// amounts the hand's need still decides.
	t.Run("mana_pick takes the larger amount", func(t *testing.T) {
		blue := spell(cardID(40), 0, "Blue Spell", "{U}{U}{U}")
		v := newView([]protocol.PlayerView{newSeat(0, withHand(blue)), newSeat(1)},
			withChoice(protocol.PendingChoiceView{
				ID: choiceID, Kind: "mana_pick", Chooser: seatID(0).String(), Reason: "Nyx Lotus",
				ColorOptions: []string{"U", "G"}, ColorAmounts: map[string]int{"U": 1, "G": 4},
			}))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "add {U}", map[string]any{"color": "U"}),
			choiceMove(t, 0, choiceID, "add {G}{G}{G}{G}", map[string]any{"color": "G"}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "add {G}{G}{G}{G}" {
			t.Fatalf("chose %q", got)
		}

		plain := newView([]protocol.PlayerView{newSeat(0, withHand(blue)), newSeat(1)},
			withChoice(protocol.PendingChoiceView{
				ID: choiceID, Kind: "mana_pick", Chooser: seatID(0).String(), Reason: "Birds",
				ColorOptions: []string{"G", "U"},
			}))
		in = input(0, plain,
			choiceMove(t, 0, choiceID, "add {G}", map[string]any{"color": "G"}),
			choiceMove(t, 0, choiceID, "add {U}", map[string]any{"color": "U"}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "add {U}" {
			t.Fatalf("equal amounts: chose %q, want the colour the hand needs", got)
		}
	})
}

func TestCleanupDiscardPitchesTheWorstCard(t *testing.T) {
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	mountain := land(cardID(2), 0)
	var bf []protocol.CardView
	for i := 0; i < 6; i++ {
		bf = append(bf, land(cardID(30+i), 0))
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(dragon, mountain)), newSeat(1)},
		withBattlefield(bf...))
	in := input(0, v,
		discardMove(t, 0, "discard the dragon", cardID(1)),
		discardMove(t, 0, "discard the land", cardID(2)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard the land" {
		t.Fatalf("chose %q", got)
	}
}

// The commander is the deck's best card and it comes back when it
// dies, so a bot that treats it as one more 3/3 will leave it in the
// command zone all game.
func TestCommanderIsWorthCastingOverAnEqualBody(t *testing.T) {
	cmdr := creature(cardID(1), 0, "Commander Bear", 3, 3, commander())
	plain := creature(cardID(2), 0, "Bear", 3, 3)
	seat := newSeat(0, withHand(plain))
	seat.Command = protocol.ZoneView{Kind: "command", Count: 1, Cards: []protocol.CardView{cmdr}}
	var bf []protocol.CardView
	for i := 0; i < 5; i++ {
		bf = append(bf, land(cardID(30+i), 0))
	}
	v := newView([]protocol.PlayerView{seat, newSeat(1)}, withBattlefield(bf...))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, cardID(2), "Cast Bear"),
		castMove(t, 0, cardID(1), "Cast Commander Bear"),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Commander Bear" {
		t.Fatalf("chose %q", got)
	}
}

// Ganging up on an attacker that is already blocked has to clear a
// higher bar than the first block did: the damage is already stopped,
// so the only thing left to buy is the attacker's death.
func TestGangBlockingNeedsToChangeTheOutcome(t *testing.T) {
	attacker := creature(cardID(20), 1, "Giant", 4, 4, attacking(0))
	firstBlocker := creature(cardID(10), 0, "Wall", 0, 5, blocking(cardID(20)))
	spare := creature(cardID(11), 0, "Bear", 2, 2)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(attacker, firstBlocker, spare),
		withTurn(5, 1, "declare_blockers"),
	)
	in := input(0, v,
		blockMove(t, 0, cardID(11), cardID(20)),
		// A second, equally pointless body to make this a choice.
		blockMove(t, 0, cardID(10), cardID(20)),
	)
	if d := decide(t, heuristic.New(), in); d.Index != aiseat.Decline {
		t.Fatalf("chose %q; the attacker is already stopped and the 2/2 only dies", chose(t, in, d))
	}
}
