package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// phyrexian_life_test.go — #1677's policy half. The enumerator now
// offers a spell's Phyrexian symbols paid with life; the heuristic has
// to take the mana payment whenever one is on offer, take the life
// payment when it is the only way to cast, and stop paying life below
// its floor. The whole-engine version is aiseat/phyrexian_life_test.go.

const (
	dismemberMana = "Cast Dismember → Big Threat"
	dismemberLife = "Cast Dismember paying 4 life for Phyrexian mana → Big Threat"
)

// phyrexianCast is a cast move shaped the way legal/cast.go emits one
// (#1677): `phyrexian_life` in the params, and the life on the Move's
// Cost twice — in Life, and in PhyrexianLife, the part that buys
// nothing. Zero symbols is the mana payment, which declares no cost.
func phyrexianCast(t *testing.T, seat int, id, label string, symbols int, target map[string]string) legal.Move {
	params := map[string]any{"instance_id": id, "from_zone": "hand", "targets": []map[string]string{target}}
	var cost *legal.MoveCost
	if symbols > 0 {
		params["phyrexian_life"] = symbols
		cost = &legal.MoveCost{Life: 2 * symbols, PhyrexianLife: 2 * symbols}
	}
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindCast,
		Label: label, Source: uuid.MustParse(id), Cost: cost,
		Params: mustJSON(t, params),
	}
}

// dismemberBoard is the bot at `life` holding a Dismember, facing an
// opponent's big creature — a removal target the policy wants dead.
func dismemberBoard(life int) protocol.GameView {
	return newView(
		[]protocol.PlayerView{
			newSeat(0, withLife(life), withHand(spell(cardID(1), 0, "Dismember", "{1}{B/P}{B/P}"))),
			newSeat(1),
		},
		withBattlefield(creature(cardID(2), 1, "Big Threat", 6, 6)),
	)
}

// TestHeuristicPrefersManaToPhyrexianLife — with the mana payment on
// offer the bot takes it, at a comfortable total and at a huge one.
// Life spent on a Phyrexian symbol buys the same spell, so it is a pure
// price; before costValue withheld the LifePayoff proxy from it, the
// life payment scored ABOVE the mana one at 40 life.
func TestHeuristicPrefersManaToPhyrexianLife(t *testing.T) {
	for _, life := range []int{20, 40} {
		in := input(0, dismemberBoard(life),
			passMove(0),
			phyrexianCast(t, 0, cardID(1), dismemberLife, 2, cardTarget(cardID(2))),
			phyrexianCast(t, 0, cardID(1), dismemberMana, 0, cardTarget(cardID(2))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != dismemberMana {
			t.Errorf("at %d life the bot chose %q, want the mana payment", life, got)
		}
	}
}

// TestHeuristicPaysPhyrexianLifeAboveTheFloor — when life is the only
// way to cast, the bot pays it while the total stays at or above the
// floor (10), and declines once paying would take it below.
func TestHeuristicPaysPhyrexianLifeAboveTheFloor(t *testing.T) {
	for _, tc := range []struct {
		life int
		want string
	}{
		{20, dismemberLife},
		{14, dismemberLife}, // lands on 10, the floor itself
		{13, "Pass priority"},
		{5, "Pass priority"},
	} {
		in := input(0, dismemberBoard(tc.life),
			passMove(0),
			phyrexianCast(t, 0, cardID(1), dismemberLife, 2, cardTarget(cardID(2))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
			t.Errorf("at %d life the bot chose %q, want %q", tc.life, got, tc.want)
		}
	}
}
