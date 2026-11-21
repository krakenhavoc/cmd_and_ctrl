package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Safeguard — Enchantment {3}{W}{W}:
//
//	"{2}{W}: Prevent all combat damage that would be dealt by target creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the ability resolves (CR 400.7).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e310c3ab-a729-404d-944f-9b477258495c",
		Name:         "Safeguard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.ShieldAgainstTargetsRow(
			"{2}{W}: Prevent all combat damage that would be dealt by target creature this turn.",
			effects.ManaCost("{2}{W}"), effects.TargetCreature("target creature"), true)},
	})
}
