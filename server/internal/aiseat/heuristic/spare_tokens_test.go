package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spare_tokens_test.go is #4991: Sai, Master Thopterist's "{1}{U},
// Sacrifice two artifacts: Draw a card" with a pile of Thopters. Two
// Thopters cost their whole board value (2 × 2.95) against the card's
// 1.20, so the bot never drew, however many it had.

// sai is Sai, Master Thopterist with its draw row as the catalog declares
// it: `value`, one card.
func sai(id string) protocol.CardView {
	c := creature(id, 0, "Sai, Master Thopterist", 1, 4, rows("triggered", "activated"))
	c.TypeLine, c.ManaCost = "Legendary Creature — Human Artificer", "{2}{U}"
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, SacrificeLabel: "two artifacts",
		Purpose: &protocol.PurposeView{Draws: 1, Answers: &[]string{"value"}},
	}}
	return c
}

// thopter is a 1/1 colorless Thopter artifact creature token with flying.
func thopter(id string) protocol.CardView {
	c := creature(id, 0, "Thopter", 1, 1, keywords("flying"))
	c.TypeLine, c.ManaCost, c.IsToken = "Token Artifact Creature — Thopter", "", true
	return c
}

// saiBoard is Sai and n Thopters, each changed by opts, with the
// activation that sacrifices the first two (the enumerator's payment
// order: tokens first, in board order).
func saiBoard(t *testing.T, n int, step string, opts ...cardOpt) (protocol.GameView, []legal.Move) {
	t.Helper()
	cards := []protocol.CardView{sai(cardID(1))}
	for i := 0; i < n; i++ {
		c := thopter(cardID(10 + i))
		for _, o := range opts {
			o(&c)
		}
		cards = append(cards, c)
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(cards...), withTurn(7, 0, step))
	return v, []legal.Move{passMove(0), sacMove(t, 0, cardID(1), "Sai: draw (sacrificing Thopter, Thopter)", cardID(10), cardID(11))}
}

func withoutSpareTokens() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceSpareTokens = false
	return heuristic.NewWithConfig(cfg)
}

// Five Thopters after combat: two of them for a card leaves three, the
// pile SpareTokenKeep keeps, so they are spare and the bot draws. With the
// price off it passes, as it did in every arena window of the issue.
func TestSaiCashesInSpareThoptersAfterCombat(t *testing.T) {
	v, moves := saiBoard(t, 5, "postcombat_main", tapped())
	in := input(0, v, moves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != moves[1].Label {
		t.Errorf("the heuristic chose %q, want Sai's draw with five Thopters", got)
	}
	if got := chose(t, in, decide(t, withoutSpareTokens(), in)); got != "Pass priority" {
		t.Errorf("with PriceSpareTokens off the heuristic chose %q, want the pass it made before #4991", got)
	}
}

// Three Thopters: the draw would leave one, under the kept pile, so each
// still costs its board value and the bot keeps its army.
func TestSaiKeepsASmallPileOfThopters(t *testing.T) {
	v, moves := saiBoard(t, 3, "postcombat_main", tapped())
	in := input(0, v, moves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass with three Thopters", got)
	}
}

// Before combat the same five Thopters can still attack, so none is spare
// yet: they swing first, and are cashed in after combat.
func TestSaiWaitsForTheThoptersToAttack(t *testing.T) {
	v, moves := saiBoard(t, 5, "precombat_main")
	in := input(0, v, moves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q before combat, want a pass while the Thopters can attack", got)
	}
	// Summoning-sick Thopters cannot attack, so they are spare already.
	v, moves = saiBoard(t, 5, "precombat_main", sick())
	in = input(0, v, moves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != moves[1].Label {
		t.Errorf("with summoning-sick Thopters the heuristic chose %q, want Sai's draw", got)
	}
}

// A Thopter with a counter on it is no longer one of the pile: an oil
// counter leaves it a 1/1, so only the counter tells it apart.
func TestACounteredTokenIsNotSpare(t *testing.T) {
	v, moves := saiBoard(t, 5, "postcombat_main", tapped())
	in := input(0, v, moves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != moves[1].Label {
		t.Fatalf("without counters the heuristic chose %q, want Sai's draw", got)
	}
	v.Battlefield.Cards[1].Counters = map[string]int{"oil": 1}
	v.Battlefield.Cards[2].Counters = map[string]int{"oil": 1}
	in = input(0, v, moves...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass when the Thopters it would sacrifice carry counters", got)
	}
}
