package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// player_mana_ability_test.go — #4885. A player mana ability (Channel's
// "pay 1 life: add {C}" until end of turn) reaches the bot as an
// activate_mana_ability move with no card and no Source, carrying
// `player_ability` and a life Cost. The heuristic has no rule that makes
// paying life to float mana worth it, so it never takes the move; and a
// move with no Source must not crash any reader.

func playerManaMove(t *testing.T, seat int) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateManaAbility, Player: seatID(seat), Kind: legal.KindMana,
		Label:  "Channel: pay 1 life, add {C}",
		Cost:   &legal.MoveCost{Life: 1},
		Params: mustJSON(t, map[string]any{"player_ability": uuid.NewString()}),
	}
}

func TestPlayerManaAbilityIsNeverTakenOverAPass(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)})
	in := input(0, v, passMove(0), playerManaMove(t, 0), playerManaMove(t, 0))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q, want the pass", got)
	}
}

func TestPlayerManaAbilityIsPricedBelowAFreeFloat(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(land(cardID(10), 0)))
	in := input(0, v, playerManaMove(t, 0), manaMove(t, 0, cardID(10)))
	ranked := heuristic.New().Rank(context.Background(), in)
	if len(ranked) != 2 {
		t.Fatalf("ranked %d moves, want 2", len(ranked))
	}
	if ranked[0].Index == 0 {
		t.Fatalf("the life-paying float ranked above a free one: %+v", ranked)
	}
}

// With nothing else on offer the policy still answers without a panic.
func TestPlayerManaAbilityAloneDoesNotCrashTheBot(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)})
	in := input(0, v, playerManaMove(t, 0))
	if _, err := heuristic.New().Decide(context.Background(), in); err != nil {
		t.Logf("decide: %v", err)
	}
}
