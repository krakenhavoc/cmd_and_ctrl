package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Time of Need, Journeyer's Kite --------------------------------------------------------

// chkB1OpenSearch is the open library-search prompt, or nil.
func ChkB1OpenSearch(g *game.Game) *game.PendingChoice {
	var open *game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
				open = c
				return
			}
		}
	})
	return open
}

// chkB1GenericCost is a mana cost of mana value n.
func ChkB1GenericCost(n int) string {
	if n <= 0 {
		return "{0}"
	}
	return "{" + string(rune('0'+n)) + "}"
}
