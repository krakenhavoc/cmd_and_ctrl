package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// free_repeat_test.go — #2777. A heuristic seat bought Shaman en-Kor's
// free {0} shield in every window, again and again, until the CR 732
// breaker paused the table.

const enKorLabel = "{0}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature you control instead."

// enKor is a Shaman en-Kor as the wire sends it: one free row, no tap.
func enKor(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Shaman en-Kor", 1, 2)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: enKorLabel, ManaCost: "{0}",
	}}
	return c
}

func enKorMove(t *testing.T, seat int, src, target string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: enKorLabel, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{
			"source_card_id": src, "ability_index": 0,
			"targets": []map[string]string{cardTarget(target)},
		}),
	}
}

func enKorTable(t *testing.T, turn int) aiseat.Input {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(enKor(cardID(1), 0), creature(cardID(2), 0, "Grizzly Bears", 2, 2)),
		withTurn(turn, 0, "precombat_main"))
	return input(0, v, passMove(0), enKorMove(t, 0, cardID(1), cardID(2)))
}

// valueOfMove is what Rank prices in.Moves[i] at.
func valueOfMove(t *testing.T, p *heuristic.Policy, in aiseat.Input, i int) float64 {
	t.Helper()
	for _, c := range p.Rank(context.Background(), in) {
		if c.Index == i {
			return c.Value
		}
	}
	t.Fatalf("Rank priced no move %d", i)
	return 0
}

func TestTheBotDoesNotRepeatAFreeActivationInOneTurn(t *testing.T) {
	p := heuristic.New()
	in := enKorTable(t, 4)
	if got := chose(t, in, decide(t, p, in)); got != enKorLabel {
		t.Fatalf("first window: the bot chose %q; the control expects the flat price to take the free row once", got)
	}
	// The shield is in place; the next window offers the same row.
	again := enKorTable(t, 4)
	if v := valueOfMove(t, p, again, 1); v > 0 {
		t.Errorf("the second free activation in the same turn is priced %.2f, above passing (#2777)", v)
	}
	if got := chose(t, again, decide(t, p, again)); got != passMove(0).Label {
		t.Errorf("second window: the bot chose %q; a second {0} shield adds nothing (#2777)", got)
	}
	// A new turn: the shield is gone, and the row is worth one more.
	next := enKorTable(t, 5)
	if got := chose(t, next, decide(t, p, next)); got != enKorLabel {
		t.Errorf("next turn: the bot chose %q; the first free activation of a turn keeps its price", got)
	}
}

// The control: the rule is about free rows. A row that costs mana is
// bounded by the mana, and keeps its price on a second activation.
func TestTheBotStillRepeatsARowThatCostsMana(t *testing.T) {
	p := heuristic.New()
	mk := func() aiseat.Input {
		c := enKor(cardID(1), 0)
		c.ActivatedAbilities[0].ManaCost = "{1}"
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(c, creature(cardID(2), 0, "Grizzly Bears", 2, 2), land(cardID(10), 0), land(cardID(11), 0)),
			withTurn(4, 0, "precombat_main"))
		m := enKorMove(t, 0, cardID(1), cardID(2))
		m.Cost = &legal.MoveCost{Mana: "{1}"}
		return input(0, v, passMove(0), m)
	}
	first := mk()
	_ = decide(t, p, first)
	before := valueOfMove(t, heuristic.New(), first, 1)
	if after := valueOfMove(t, p, mk(), 1); after != before {
		t.Errorf("a {1} row is priced %.2f after one activation, %.2f before: only free rows are capped", after, before)
	}
}
