package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sacrifice_all_test.go — #2097: the policy cannot price a spell whose
// cost sacrifices every creature it controls (Soulblast), so it never
// casts one: not over a board it would wipe, and not over an empty
// board where the spell does nothing.

func soulblastCard(id string, all bool, sacrificed ...string) protocol.CardView {
	c := spell(id, 0, "Soulblast", "{3}{R}{R}{R}")
	c.AdditionalCost = &protocol.AdditionalCostView{
		Label: "Sacrifice all creatures you control",
		SacrificeOptions: &protocol.LegalTargetsView{
			Min: len(sacrificed), Max: len(sacrificed), All: all, Cards: sacrificed,
		},
	}
	return c
}

func soulblastMove(t *testing.T, id string, sacrificed ...string) legal.Move {
	params := map[string]any{"instance_id": id, "from_zone": "hand", "targets": []map[string]string{playerTarget(1)}}
	if len(sacrificed) > 0 {
		params["sacrifice_ids"] = sacrificed
	}
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(0), Kind: legal.KindCast,
		Label: "Cast Soulblast", Source: uuid.MustParse(id), Params: mustJSON(t, params),
	}
}

func TestTheBotNeverCastsASacrificeAllSpell(t *testing.T) {
	t.Run("over its own creatures", func(t *testing.T) {
		bears := []string{cardID(2), cardID(3)}
		sb := soulblastCard(cardID(1), true, bears...)
		v := newView(
			[]protocol.PlayerView{newSeat(0, withHand(sb)), newSeat(1)},
			withBattlefield(creature(bears[0], 0, "Bear", 2, 2), creature(bears[1], 0, "Bear", 2, 2)),
			withTurn(5, 0, "precombat_main"),
		)
		in := input(0, v, passMove(0), soulblastMove(t, cardID(1), bears...))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("the heuristic chose %q, want a pass", got)
		}
	})
	t.Run("over an empty board", func(t *testing.T) {
		sb := soulblastCard(cardID(1), true)
		v := newView(
			[]protocol.PlayerView{newSeat(0, withHand(sb)), newSeat(1)},
			withTurn(5, 0, "precombat_main"),
		)
		in := input(0, v, passMove(0), soulblastMove(t, cardID(1)))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
			t.Errorf("the heuristic chose %q, want a pass", got)
		}
	})
}
