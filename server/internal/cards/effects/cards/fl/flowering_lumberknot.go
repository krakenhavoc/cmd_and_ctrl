package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowering Lumberknot — Creature — Treefolk {3}{G}, 5/5 (Avacyn Restored,
// S82 OLD, #5570):
//
//	"This creature can't attack or block unless it's paired with a creature
//	 with soulbond."
//
// Soulbond's pair (CR 702.95b) read from both sides: the attack half is an
// attack restriction on the creature (AttackTargetRestriction.
// MustBePairedWithKeyword, judged at declaration, CR 508.1c) and the block
// half a block rule (CR 509.1b). The partner must itself have soulbond, so a
// Lumberknot paired with another Lumberknot is still stuck. Lumberknot has no
// soulbond itself, so it never pairs on its own: it needs a soulbond creature
// to pair with it.
//
// No simplification.
func init() {
	attack, block := effects.S82OldB437CantAttackOrBlockUnlessPairedWithSoulbond()
	effects.Register(effects.Spec{
		OracleID:     "97e0190d-a9a5-432d-9694-7576477bf65e",
		Name:         "Flowering Lumberknot",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{attack},
		BlockRules:   []game.BlockRule{block},
	})
}
