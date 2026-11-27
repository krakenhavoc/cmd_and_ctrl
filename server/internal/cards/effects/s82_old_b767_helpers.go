package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b767_helpers.go — shared pieces of S82 OLD batch 767 (#5628):
// Tempest creatures with keywords.

// CanBlockAsThoughItHadShadow — "This creature can block creatures with shadow
// as though it had shadow" (Wall of Diffusion; CR 702.28b). The waiver belongs
// to the permanent that prints it and to the attackers the queries name; it
// does not gain shadow, so a card that counts creatures with shadow does not
// see it.
func CanBlockAsThoughItHadShadow(queries ...game.PermanentQuery) game.BlockRule {
	return game.BlockRule{ShadowAgainst: queries}
}

// CreaturesWithShadow is the query "creatures with shadow".
func CreaturesWithShadow() game.PermanentQuery {
	return game.PermanentQuery{Types: []string{"creature"}, Keyword: "shadow"}
}
