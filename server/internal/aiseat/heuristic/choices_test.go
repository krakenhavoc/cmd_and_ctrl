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

	// Owner decision (2026-09-17): an "any color" pick lists all five
	// colours with the commander's identity first. With nothing in hand
	// asking for a colour, the bot keeps the first — an identity
	// colour — rather than wandering off it.
	t.Run("mana_pick with no hand need takes the identity colour listed first", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withChoice(protocol.PendingChoiceView{
				ID: choiceID, Kind: "mana_pick", Chooser: seatID(0).String(), Reason: "Birds of Paradise",
				ColorOptions: []string{"G", "W", "U", "B", "R"},
			}))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "add {G}", map[string]any{"color": "G"}),
			choiceMove(t, 0, choiceID, "add {W}", map[string]any{"color": "W"}),
			choiceMove(t, 0, choiceID, "add {U}", map[string]any{"color": "U"}),
			choiceMove(t, 0, choiceID, "add {B}", map[string]any{"color": "B"}),
			choiceMove(t, 0, choiceID, "add {R}", map[string]any{"color": "R"}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "add {G}" {
			t.Fatalf("chose %q, want the identity colour listed first", got)
		}
	})
}

func TestCoinCallUsesChoiceIDAndStopsAtDeclaredCeiling(t *testing.T) {
	const choiceID = "00000000-0000-4000-8000-000000000001" // tails
	base := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withChoice(protocol.PendingChoiceView{
		ID: choiceID, Kind: "coin_call", Chooser: seatID(0).String(),
	}))
	callMoves := []legal.Move{
		choiceMove(t, 0, choiceID, "call heads", map[string]any{"call": "heads"}),
		choiceMove(t, 0, choiceID, "call tails", map[string]any{"call": "tails"}),
	}
	in := input(0, base, callMoves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "call tails" {
		t.Fatalf("coin call = %q, want choice-ID tails", got)
	}

	stopView := base
	stopView.PendingChoices[0].AllowStop = true
	stopView.PendingChoices[0].MaxUsefulWins = 3
	stopView.PendingChoices[0].Wins = 3
	stop := choiceMove(t, 0, choiceID, "stop flipping", map[string]any{"call": "stop"})
	in = input(0, stopView, append(callMoves, stop)...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "stop flipping" {
		t.Fatalf("coin ceiling = %q, want stop", got)
	}
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

// --- #798: a choose_cards prompt over the bot's own hand -------------

// sixLands is a settled mana base: with this much on the battlefield
// the bot is past Config.LandsWanted, so another land in hand is the
// cheapest card it holds. That is the same board
// TestCleanupDiscardPitchesTheWorstCard uses, and it is what makes
// "worst card" mean the same thing on both discard paths.
func sixLands() []protocol.CardView {
	var bf []protocol.CardView
	for i := 0; i < 6; i++ {
		bf = append(bf, land(cardID(30+i), 0))
	}
	return bf
}

// handChoice is a choose_cards prompt over seat 0's own hand — the
// shape #797 gives every effect discard, loot and rummage.
func handChoice(id, reason string, lo, hi int, opts ...protocol.CardView) protocol.PendingChoiceView {
	return protocol.PendingChoiceView{
		ID: id, Kind: "choose_cards", Chooser: seatID(0).String(),
		FromPlayer: seatID(0).String(), Reason: reason,
		Options: opts, ChooseMin: lo, ChooseMax: hi, Count: hi,
	}
}

// #798. Before this, every choose_cards answer scored the same and the
// bot took the first one the enumerator offered — which is hand order,
// so a Mind Rotted bot pitched whatever it happened to be holding
// first. The answer is scored by what it KEEPS, so the cards it names
// are the ones it wants least. The winning answer is listed LAST in
// every case here: a test that passed on the enumerator's ordering
// would be testing nothing.
func TestEffectDiscardNamesTheWorstCardsNotTheFirstOnes(t *testing.T) {
	const choiceID = "choice-discard"
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	mountain := land(cardID(3), 0)

	t.Run("Mind Rot discards the two worst", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(dragon, bear, mountain)), newSeat(1)},
			withBattlefield(sixLands()...),
			withChoice(handChoice(choiceID, "Mind Rot — discard 2 cards", 2, 2, dragon, bear, mountain)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "pitch dragon + bear", map[string]any{"card_ids": []string{cardID(1), cardID(2)}}),
			choiceMove(t, 0, choiceID, "pitch dragon + mountain", map[string]any{"card_ids": []string{cardID(1), cardID(3)}}),
			choiceMove(t, 0, choiceID, "pitch bear + mountain", map[string]any{"card_ids": []string{cardID(2), cardID(3)}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch bear + mountain" {
			t.Fatalf("chose %q, want the answer that keeps the Dragon", got)
		}
	})

	// "Discard up to two cards" — the bot owes nothing, so it keeps
	// everything. Keeping a card is never worth less than nothing, so
	// the smallest legal answer wins without a rule of its own.
	t.Run("up to two discards as few as it must", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(dragon, mountain)), newSeat(1)},
			withBattlefield(sixLands()...),
			withChoice(handChoice(choiceID, "Rummage — discard up to 2 cards", 0, 2, dragon, mountain)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "pitch both", map[string]any{"card_ids": []string{cardID(1), cardID(3)}}),
			choiceMove(t, 0, choiceID, "pitch the dragon", map[string]any{"card_ids": []string{cardID(1)}}),
			choiceMove(t, 0, choiceID, "pitch the mountain", map[string]any{"card_ids": []string{cardID(3)}}),
			choiceMove(t, 0, choiceID, "pitch nothing", map[string]any{"card_ids": []string{}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch nothing" {
			t.Fatalf("chose %q, want the smallest legal answer", got)
		}
	})

	// A loot draws first and then discards, so the drawn card is in
	// the hand the prompt is built from and the count is fixed: the
	// only question is which of the two the bot keeps.
	t.Run("looting keeps the better card", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(mountain, dragon)), newSeat(1)},
			withBattlefield(sixLands()...),
			withChoice(handChoice(choiceID, "Faithless Looting — discard a card", 1, 1, mountain, dragon)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "pitch the dragon", map[string]any{"card_ids": []string{cardID(1)}}),
			choiceMove(t, 0, choiceID, "pitch the mountain", map[string]any{"card_ids": []string{cardID(3)}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch the mountain" {
			t.Fatalf("chose %q, want the answer that keeps the Dragon", got)
		}
	})

	// A rummage discards and then draws — the prompt is the same
	// shape, over the hand as it stands before the draw.
	t.Run("rummage names the worst card it holds", func(t *testing.T) {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(dragon, bear, mountain)), newSeat(1)},
			withBattlefield(sixLands()...),
			withChoice(handChoice(choiceID, "Syphon Mind — discard a card", 1, 1, dragon, bear, mountain)))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "pitch the dragon", map[string]any{"card_ids": []string{cardID(1)}}),
			choiceMove(t, 0, choiceID, "pitch the bear", map[string]any{"card_ids": []string{cardID(2)}}),
			choiceMove(t, 0, choiceID, "pitch the mountain", map[string]any{"card_ids": []string{cardID(3)}}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch the mountain" {
			t.Fatalf("chose %q", got)
		}
	})
}

// The rule is about cards the bot is being asked to GIVE UP, which is
// what a candidate in its own hand means. A choose_cards over anything
// else — a Ward sacrifice, a library pick, a reveal — carries no such
// promise on the wire, so it keeps the flat score and the enumerator's
// order decides. Naming the bot's best permanent because "keeping" the
// others scored higher would be a worse bug than the one #798 fixed.
func TestChooseCardsOverOtherZonesKeepsTheEnumeratorsOrder(t *testing.T) {
	const choiceID = "choice-bf"
	dragon := creature(cardID(1), 0, "Dragon", 6, 6)
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(dragon, bear),
		withChoice(handChoice(choiceID, "Ward — choose a creature to sacrifice", 1, 1, dragon, bear)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "name the dragon", map[string]any{"card_ids": []string{cardID(1)}}),
		choiceMove(t, 0, choiceID, "name the bear", map[string]any{"card_ids": []string{cardID(2)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "name the dragon" {
		t.Fatalf("chose %q, want the enumerator's first answer", got)
	}
}

// The policy carries no randomness and reads no map in an order that
// could vary, so the same window must answer the same way every time —
// including when two answers are worth exactly the same and the tie
// falls to the enumerator's order. #775 keyed the RNG the rest of the
// bot stack draws from; this is the part of it that has none.
func TestDiscardRankingIsDeterministicAcrossRepeatedDecisions(t *testing.T) {
	const choiceID = "choice-tie"
	first := land(cardID(1), 0)
	second := land(cardID(2), 0)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(first, second)), newSeat(1)},
		withBattlefield(sixLands()...),
		withChoice(handChoice(choiceID, "Mind Rot — discard a card", 1, 1, first, second)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "pitch the first", map[string]any{"card_ids": []string{cardID(1)}}),
		choiceMove(t, 0, choiceID, "pitch the second", map[string]any{"card_ids": []string{cardID(2)}}),
	)
	pol := heuristic.New()
	for i := 0; i < 50; i++ {
		if got := chose(t, in, decide(t, pol, in)); got != "pitch the first" {
			t.Fatalf("run %d chose %q; two equal answers must always tie to the enumerator's order", i, got)
		}
	}
}
