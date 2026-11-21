package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b197_helpers.go — shared by the GAP batch 197 cards (#5948).

// CreaturesThatDiedThisTurnAmount is the amount of "this spell costs {1}
// less to cast for each creature that died this turn" (Blood for the Blood
// God!): the table-wide tally of creatures that died this turn, tokens
// included, read at CR 601.2f. A query with no game reads zero.
func CreaturesThatDiedThisTurnAmount(q game.CostQuery) int {
	if q.Game != nil {
		return B11CreaturesDiedThisTurn(q.Game)
	}
	return 0
}
