package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Standard Bearer — Creature — Human Flagbearer {1}{W}, 1/1 (Apocalypse,
// #5614):
//
//	"While an opponent is choosing targets as part of casting a spell they
//	 control or activating an ability they control, that player must choose at
//	 least one Flagbearer on the battlefield if able."
//
// Coalition Honor Guard's Flagbearer rule (#6236): a self static bound at the
// cast and activation gates, with the Flagbearer subtype printed on the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d6e9716-8c3c-443d-afd8-68d0d6c469a5",
		Name:         "Standard Bearer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.FlagbearerTargetRule()},
	})
}
