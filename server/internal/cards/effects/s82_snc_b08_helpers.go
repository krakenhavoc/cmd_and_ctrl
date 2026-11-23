package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_snc_b08_helpers.go — shared reads for the Streets of New Capenna
// counters-matter batch (S82, #5491).

// s82SncHasACounter reports whether the card carries a counter of any kind.
func S82SncHasACounter(c game.Card) bool {
	return B41HasACounter()(nil, [16]byte{}, c)
}
