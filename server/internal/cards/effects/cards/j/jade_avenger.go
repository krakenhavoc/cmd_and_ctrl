package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jade Avenger — Creature — Frog Samurai {1}{G}, 2/2 (Modern Horizons 2,
// #5501):
//
//	"Bushido 2 (Whenever this creature blocks or becomes blocked, it gets
//	 +2/+2 until end of turn.)"
//
// Samurai Enforcers' bushido 2: the shared Bushido trigger (CR 702.45a),
// one per block or becoming blocked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "177248b7-f3fd-42f8-b325-2625a2bb28ff",
		Name:         "Jade Avenger",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Bushido(2, "Jade Avenger")},
	})
}
