package mcpseat

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #3870: a move whose cost mills says how many cards, as one that pays
// energy does.
func TestMoveLineSaysTheCardsItMills(t *testing.T) {
	v := &protocol.GameView{}
	for _, tc := range []struct {
		n    int
		want string
	}{
		{1, "mills a card from your library"},
		{4, "mills 4 cards from your library"},
	} {
		m := legal.Move{Label: "Rot Farm Skeleton: …", Cost: &legal.MoveCost{Mill: tc.n}}
		if got := moveLine(0, m, v, "", newNameWrapper(v)); !strings.Contains(got, tc.want) {
			t.Errorf("mill %d: move line %q does not say %q", tc.n, got, tc.want)
		}
	}
}
