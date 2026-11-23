package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caverns of Despair — World Enchantment {2}{R}{R} (EDHREC rank
// 21037):
//
//	"No more than two creatures can attack each combat.
//	 No more than two creatures can block each combat."
//
// Silent Arbiter's lines with a bound of two (#1507). Both are exact.
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1034a02-36cf-4586-a001-9dc3fb76e904",
		Name:         "Caverns of Despair",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.NoMoreThanNCanAttackEachCombat(2)},
		BlockRules:   []game.BlockRule{effects.NoMoreThanNCanBlockEachCombat(2)},
	})
}
