package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cho-Arrim Bruiser — Creature — Ogre Rebel {5}{W}, 3/4 (Mercadian Masques, S82 OLD batch 687, #5619):
//
//	"Whenever this creature attacks, you may tap up to two target creatures."
//
// Eddymurk Crab's up-to-two creature clause (any player's creatures, a target gone by
// resolution is skipped, CR 608.2b) on an attack trigger, optional at resolution
// (Batterhorn's optional targeted trigger). Choosing no target does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3db407ec-da4b-4b25-b613-e45c164ffe86",
		Name:         "Cho-Arrim Bruiser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WheneverThisAttacks("Cho-Arrim Bruiser — you may tap up to two target creatures", effects.TapEachLegalTarget),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2)),
				"Cho-Arrim Bruiser — tap the chosen creatures?"),
		},
	})
}
