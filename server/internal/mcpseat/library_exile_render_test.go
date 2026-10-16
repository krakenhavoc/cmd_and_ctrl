package mcpseat

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #3966: a move whose cost exiles the top of the library says how many
// cards, as one that mills does.
func TestMoveLineSaysTheLibraryCardsItExiles(t *testing.T) {
	v := &protocol.GameView{}
	for _, tc := range []struct {
		n    int
		want string
	}{
		{1, "exiles the top card of your library"},
		{4, "exiles the top 4 cards of your library"},
	} {
		m := legal.Move{Label: "Seasoned Tactician: …", Cost: &legal.MoveCost{LibraryExile: tc.n}}
		if got := moveLine(0, m, v, "", newNameWrapper(v)); !strings.Contains(got, tc.want) {
			t.Errorf("exile %d: move line %q does not say %q", tc.n, got, tc.want)
		}
	}
}
