package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kinsbaile Cavalier — Creature — Kithkin Knight {3}{W}, 2/2:
//
//	"Knight creatures you control have double strike."
//
// Knight Exemplar's tribal keyword grant without "other": the Cavalier is
// itself a Knight, so it has double strike too, and a Knight loses it the
// moment the Cavalier leaves.
//
// No simplification.
func init() {
	knights := effects.TribeFilter{Tribes: []string{"Knight"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "e21ed0d0-26b1-4ffe-971a-babf37f4492c",
		Name:         "Kinsbaile Cavalier",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(knights, "double strike"),
		},
	})
}
