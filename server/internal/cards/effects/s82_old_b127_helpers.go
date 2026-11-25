package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// SomeGraveyardHasAtLeast reports whether any one player's graveyard holds n
// or more cards: Swimmer in Nightmares' "ten or more cards in a single
// graveyard" (Theros Beyond Death, S82 OLD, #5514). Read live off the table,
// so a static that asks it applies and stops applying as graveyards change.
func SomeGraveyardHasAtLeast(g *game.Game, n int) bool {
	if g == nil {
		return false
	}
	for _, p := range g.Seats {
		if p != nil && p.Graveyard != nil && p.Graveyard.Size() >= n {
			return true
		}
	}
	return false
}
