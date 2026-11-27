package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Jailer — Creature — Goblin {1}{R}, 2/2 (Planeshift, S82 OLD b644, #5615):
//
//	"This creature can't attack if defending player controls an untapped creature
//	 with power 2 or less."
//
// The opposite of Sea Serpent's "unless" (CantAttackIfDefendingPlayerControls, CR
// 508.1c): each target's defending player (CR 508.5) is asked at declaration, so in
// Commander it may attack the opponents who control no untapped creature with power 2
// or less and nobody else. A creature's power is read as it is now, layers included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "423659f1-1b48-4a2b-a1dc-f39fd62c6b6a",
		Name:         "Mogg Jailer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CantAttackIfDefendingPlayerControls(game.PermanentQuery{Types: []string{"creature"}, Untapped: true}.WithPowerAtMost(2)),
		},
	})
}
