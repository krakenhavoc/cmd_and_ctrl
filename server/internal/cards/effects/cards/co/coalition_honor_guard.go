package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coalition Honor Guard — {3}{W} Creature — Human Flagbearer 2/4 (Eternal
// Masters, S82 OLD, #5542):
//
//	"Flagbearer (While an opponent is choosing targets as part of casting a
//	 spell they control or activating an ability they control, that player
//	 must choose at least one Flagbearer on the battlefield if able.)"
//
// The Flagbearer seam (#6236): a self static that marks the permanent, bound at
// the cast and activation gates. The Flagbearer subtype is printed on the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c95cebf7-c28c-4453-b5e4-e895d621f290",
		Name:         "Coalition Honor Guard",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.FlagbearerTargetRule()},
	})
}
