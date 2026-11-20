package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "da67c329-ca12-45fc-a9cd-8fb6f27501a9",
		Name:            "Demon's Jester",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{disSelfPumpWhileHellbent(2, 1)},
	})
}
