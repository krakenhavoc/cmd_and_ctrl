package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// forage_mana_test.go — #5397. A mana ability whose cost forages
// (Thornvault Forager's "{T}, Forage: Add two mana in any combination of
// colors") reaches the bot as an activate_mana_ability move carrying
// `cost_branch`. Floating mana buys the bot nothing (its casts auto-tap,
// and the auto-tapper never forages), so the move is priced below a
// plain float and never taken over a pass.

func forageManaMove(t *testing.T, seat int, id string, branch int) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateManaAbility, Player: seatID(seat), Kind: legal.KindMana,
		Label: "Thornvault Forager: forage, add two mana", Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"card_id": id, "ability_index": 1, "cost_branch": branch}),
	}
}

func TestForageManaMoveIsNeverTakenOverAPass(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(land(cardID(10), 0)))
	in := input(0, v, passMove(0), forageManaMove(t, 0, cardID(10), 0), forageManaMove(t, 0, cardID(10), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q, want the pass", got)
	}
}

func TestForageManaMoveIsPricedBelowAPlainFloat(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(land(cardID(10), 0)))
	in := input(0, v, forageManaMove(t, 0, cardID(10), 1), manaMove(t, 0, cardID(10)))
	ranked := heuristic.New().Rank(context.Background(), in)
	if len(ranked) != 2 {
		t.Fatalf("ranked %d moves, want 2", len(ranked))
	}
	if ranked[0].Index == 0 {
		t.Fatalf("the forage float ranked above a plain one: %+v", ranked)
	}
}
