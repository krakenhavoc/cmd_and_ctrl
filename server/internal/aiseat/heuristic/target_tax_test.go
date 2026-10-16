package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// target_tax_test.go — #3860's policy half. Terror of the Peaks charges
// life for a spell that targets it (CR 601.2f), and the enumerator
// states that on the move: Cost.Life, with the part that is the tax in
// Cost.TargetTaxLife. Like the life that pays a Phyrexian symbol, that
// life buys nothing -- it is the same spell at a worse price for naming
// the taxed permanent -- so the policy must read it as a pure price and
// take an equally good, untaxed target when it has one.

const (
	boltAtTaxed = "Cast Removal paying 3 life for what it targets → Taxed Dragon"
	boltAtFree  = "Cast Removal → Free Dragon"
)

// taxedCast is a cast move shaped the way legal/cast.go emits one at a
// taxed target (#3860): the life on the Cost twice, in Life and in
// TargetTaxLife.
func taxedCast(t *testing.T, seat int, id, label string, tax int, target map[string]string) legal.Move {
	params := map[string]any{"instance_id": id, "from_zone": "hand", "targets": []map[string]string{target}}
	var cost *legal.MoveCost
	if tax > 0 {
		cost = &legal.MoveCost{Life: tax, TargetTaxLife: tax}
	}
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindCast,
		Label: label, Source: uuid.MustParse(id), Cost: cost,
		Params: mustJSON(t, params),
	}
}

// twoDragons is the bot at `life` holding a removal spell, facing two
// equal opposing creatures.
func twoDragons(life int) protocol.GameView {
	return newView(
		[]protocol.PlayerView{
			newSeat(0, withLife(life), withHand(spell(cardID(1), 0, "Removal", "{R}"))),
			newSeat(1),
		},
		withBattlefield(
			creature(cardID(2), 1, "Taxed Dragon", 5, 4),
			creature(cardID(3), 1, "Free Dragon", 5, 4),
		),
	)
}

func TestHeuristicPrefersTheUntaxedTargetOfTwoEqualOnes(t *testing.T) {
	for _, life := range []int{20, 40} {
		in := input(0, twoDragons(life),
			passMove(0),
			taxedCast(t, 0, cardID(1), boltAtTaxed, 3, cardTarget(cardID(2))),
			taxedCast(t, 0, cardID(1), boltAtFree, 0, cardTarget(cardID(3))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != boltAtFree {
			t.Errorf("at %d life the bot chose %q, want the untaxed target", life, got)
		}
	}
}

// With nothing else to point it at the tax is just another life cost: paid
// while it keeps the bot alive, refused as a suicide when it would not
// (costValue's refusal, which Phyrexian life's higher floor is a refinement
// of, not a replacement for).
func TestHeuristicPaysTheTargetTaxOnlyWhileItSurvives(t *testing.T) {
	for _, tc := range []struct {
		life int
		want string
	}{
		{40, boltAtTaxed},
		{3, "Pass priority"}, // paying 3 of 3 is its last life
	} {
		in := input(0, twoDragons(tc.life),
			passMove(0),
			taxedCast(t, 0, cardID(1), boltAtTaxed, 3, cardTarget(cardID(2))),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
			t.Errorf("at %d life the bot chose %q, want %q", tc.life, got, tc.want)
		}
	}
}
