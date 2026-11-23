package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demon's Jester — Creature — Imp {3}{B}, 2/2 (Dissension, #5598):
//
//	"Flying
//	 Hellbent — This creature gets +2/+1 as long as you have no cards in hand."
//
// Hellbent is disSelfPumpWhileHellbent, a layer 7c static re-read on
// every change of hand size.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da67c329-ca12-45fc-a9cd-8fb6f27501a9",
		Name:            "Demon's Jester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.DisSelfPumpWhileHellbent(2, 1)},
	})
}
