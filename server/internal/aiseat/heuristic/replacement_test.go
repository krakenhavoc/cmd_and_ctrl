package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// replacement_test.go — #2390's audit of the "may" replacements the
// default yes used to answer: CR 903.9b's commander question, the
// replacements that keep the yes (Library of Leng, Moonlit Meditation),
// a shockland's life payment, and a Clone's copy.

const (
	optionalChoiceID = "00000000-0000-4000-8000-000000012390"
	payLifeChoiceID  = "00000000-0000-4000-8000-000000022390"
	copyChoiceID     = "00000000-0000-4000-8000-000000032390"
)

func optionalReplacementDecision(t *testing.T, c protocol.PendingChoiceView) string {
	t.Helper()
	c.ID, c.Kind, c.Chooser, c.FromPlayer, c.Count = optionalChoiceID, "optional_replacement", seatID(0).String(), seatID(0).String(), 1
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withTurn(5, 1, "precombat_main"), withChoice(c))
	in := input(0, v, yesNoMoves(t, optionalChoiceID)...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

// TestCommanderBouncedToHandStaysInHand — CR 903.9b: headed for a hand
// (playable_from_zone), the commander stays there and is cast without
// the tax; headed for a library, it goes home.
func TestCommanderBouncedToHandStaysInHand(t *testing.T) {
	cmd := cardID(999)
	if got := optionalReplacementDecision(t, protocol.PendingChoiceView{
		Source: cmd, Reason: "Send commander to command zone instead?", PlayableFromZone: true,
	}); got != "no" {
		t.Errorf("headed for a hand: chose %q, want no", got)
	}
	if got := optionalReplacementDecision(t, protocol.PendingChoiceView{
		Source: cmd, Reason: "Send commander to command zone instead?",
	}); got != "yes" {
		t.Errorf("headed for a library: chose %q, want yes", got)
	}
}

// TestOtherOptionalReplacementsKeepTheYes — Library of Leng and Moonlit
// Meditation carry nothing the bot can price, and the audit keeps their
// yes (replacement.go says why).
func TestOtherOptionalReplacementsKeepTheYes(t *testing.T) {
	for _, reason := range []string{
		"Put the discarded card on top of your library instead?",
		"Moonlit Meditation — create copies of the enchanted permanent instead?",
	} {
		if got := optionalReplacementDecision(t, protocol.PendingChoiceView{Reason: reason}); got != "yes" {
			t.Errorf("%q: chose %q, want yes", reason, got)
		}
	}
}

// --- entry_pay_life ------------------------------------------------

func payLifeMoves(t *testing.T, life int) []legal.Move {
	t.Helper()
	pay := choiceMove(t, 0, payLifeChoiceID, "pay", map[string]any{"apply": true})
	if life > 0 {
		pay.Cost = &legal.MoveCost{Life: life}
	}
	return []legal.Move{pay, choiceMove(t, 0, payLifeChoiceID, "enter tapped", map[string]any{"apply": false})}
}

func payLifeDecision(t *testing.T, active int, step string, life, price int, hand []protocol.CardView, board []protocol.CardView) string {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0, withLife(life), withHand(hand...)), newSeat(1)},
		withTurn(3, active, step), withBattlefield(board...),
		withChoice(protocol.PendingChoiceView{
			ID: payLifeChoiceID, Kind: "entry_pay_life", Chooser: seatID(0).String(), FromPlayer: seatID(0).String(),
			Count: 1, Source: cardID(50), PayCost: "2 life",
		}))
	in := input(0, v, payLifeMoves(t, price)...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func oneDrop(id string) protocol.CardView {
	return creature(id, 0, "Goblin", 1, 1, func(c *protocol.CardView) { c.ManaCost = "{R}" })
}

func fiveDrop(id string) protocol.CardView {
	return creature(id, 0, "Dragon", 5, 5, func(c *protocol.CardView) { c.ManaCost = "{4}{R}" })
}

func TestShocklandPaysWhenTheManaIsSpentThisTurn(t *testing.T) {
	if got := payLifeDecision(t, 0, "precombat_main", 40, 2, []protocol.CardView{oneDrop(cardID(20))}, nil); got != "pay" {
		t.Errorf("turn-one shockland with a one-drop in hand: chose %q, want pay", got)
	}
	// An instant it can cast before the land would untap.
	bolt := spell(cardID(21), 0, "Bolt", "{R}")
	if got := payLifeDecision(t, 1, "end", 40, 2, []protocol.CardView{bolt}, nil); got != "pay" {
		t.Errorf("an opponent's end step with an instant in hand: chose %q, want pay", got)
	}
}

func TestShocklandEntersTappedWhenTheManaWouldGoUnused(t *testing.T) {
	for _, tc := range []struct {
		name   string
		active int
		step   string
		hand   []protocol.CardView
	}{
		{"nothing castable", 0, "precombat_main", []protocol.CardView{fiveDrop(cardID(20))}},
		{"an empty hand", 0, "precombat_main", nil},
		{"an opponent's turn, sorcery-speed hand", 1, "precombat_main", []protocol.CardView{oneDrop(cardID(20))}},
		{"its own end step", 0, "end", []protocol.CardView{oneDrop(cardID(20))}},
	} {
		if got := payLifeDecision(t, tc.active, tc.step, 40, 2, tc.hand, nil); got != "enter tapped" {
			t.Errorf("%s: chose %q, want enter tapped", tc.name, got)
		}
	}
}

func TestShocklandDoesNotPayLifeItCannotAfford(t *testing.T) {
	hand := []protocol.CardView{oneDrop(cardID(20))}
	if got := payLifeDecision(t, 0, "precombat_main", 2, 2, hand, nil); got != "enter tapped" {
		t.Errorf("at 2 life, pay 2: chose %q, want enter tapped", got)
	}
	if got := payLifeDecision(t, 0, "precombat_main", 9, 2, hand, nil); got != "enter tapped" {
		t.Errorf("at 9 life, pay 2 for a one-drop: chose %q, want enter tapped", got)
	}
	if got := payLifeDecision(t, 0, "precombat_main", 40, 0, hand, nil); got != "enter tapped" {
		t.Errorf("a pay branch with no price on it: chose %q, want enter tapped", got)
	}
}

// --- copy_target ---------------------------------------------------

func copyDecision(t *testing.T, board []protocol.CardView, ids ...string) string {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withTurn(5, 0, "precombat_main"), withBattlefield(board...),
		withChoice(protocol.PendingChoiceView{
			ID: copyChoiceID, Kind: "copy_target", Chooser: seatID(0).String(), FromPlayer: seatID(0).String(), Count: 1,
			Source: cardID(50),
		}))
	moves := []legal.Move{choiceMove(t, 0, copyChoiceID, "don't copy anything", map[string]any{"card_ids": []string{}})}
	for _, id := range ids {
		moves = append(moves, choiceMove(t, 0, copyChoiceID, "copy "+id, map[string]any{"card_ids": []string{id}}))
	}
	in := input(0, v, moves...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func TestCloneCopiesTheBestCreature(t *testing.T) {
	dragon := creature(cardID(1), 1, "Dragon", 5, 5, keywords("flying"))
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	if got := copyDecision(t, []protocol.CardView{dragon, bear}, bear.InstanceID, dragon.InstanceID); got != "copy "+dragon.InstanceID {
		t.Errorf("chose %q, want the dragon", got)
	}
	if got := copyDecision(t, []protocol.CardView{bear}, bear.InstanceID); got != "copy "+bear.InstanceID {
		t.Errorf("a Clone that copies nothing dies: chose %q, want the bear", got)
	}
}

func TestCloneSkipsALegendItAlreadyHas(t *testing.T) {
	legend := creature(cardID(1), 0, "Atraxa", 4, 4, keywords("flying", "vigilance", "deathtouch", "lifelink"), func(c *protocol.CardView) {
		c.TypeLine = "Legendary Creature — Phyrexian Angel Horror"
	})
	bear := creature(cardID(2), 1, "Bear", 2, 2)
	if got := copyDecision(t, []protocol.CardView{legend, bear}, legend.InstanceID, bear.InstanceID); got != "copy "+bear.InstanceID {
		t.Errorf("chose %q, want the bear over a second Atraxa the legend rule would bin", got)
	}
	theirs := legend
	theirs.Owner, theirs.Controller = seatID(1).String(), seatID(1).String()
	if got := copyDecision(t, []protocol.CardView{theirs, bear}, theirs.InstanceID, bear.InstanceID); got != "copy "+theirs.InstanceID {
		t.Errorf("an opponent's legend: chose %q, want Atraxa", got)
	}
}
