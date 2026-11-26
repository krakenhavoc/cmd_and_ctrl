package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravepurge — Instant {2}{B} (Dragons of Tarkir):
//
//	"Put any number of target creature cards from your graveyard on top of your
//	 library.
//	 Draw a card."
//
// Any number of targets, so it can be cast with none; the cards still in the
// graveyard as it resolves go on top in the order chosen and the draw takes the
// last one put there (CR 608.2c, 401.4). Shares Forever Young's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc94462f-a207-4fdb-8704-aa979c244cab",
		Name:         "Gravepurge",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("any number of target creature cards from your graveyard",
			effects.Creature(), effects.YouOwn()).WithCount(0, 0),
		Purpose:   game.Purpose{Draws: 1},
		OnResolve: effects.TuckTargetCreatureCardsOnTopThenDraw(),
	})
}
