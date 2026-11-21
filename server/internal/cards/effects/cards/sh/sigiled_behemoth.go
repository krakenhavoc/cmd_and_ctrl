package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigiled Behemoth — Creature — Beast {4}{G}{W}, 5/4 (Alara Reborn,
// #5586):
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Exalted is the canonical keyword (CR 702.83a, game/exalted.go): the
// engine derives the trigger from the printed keyword (Akrasan Squire).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "277ff122-2d6e-441e-9ec1-980cdb393982",
		Name:            "Sigiled Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
	})
}
